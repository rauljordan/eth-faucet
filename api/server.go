package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/prestonvanloon/go-recaptcha"
	faucetpb "github.com/rauljordan/goerli-faucet/api/proto/faucet"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	ipLimit          = 5 // 5 IP addresses per goerli address allowed.
	txGasLimit       = 40000
	fundingAmountWei = "32500000000000000000" // 32.5 ETH in Wei.
)

var (
	fundingAmount *big.Int
	funded        = make(map[string]bool)
	ipCounter     = make(map[string]int)
	fundingLock   sync.Mutex
	pruneDuration = time.Hour * 4
	_             = faucetpb.FaucetServer(&Server{})
)

// Server implementing the faucet gRPC service.
type Server struct {
	faucetpb.UnimplementedFaucetServer
	r           recaptcha.Recaptcha
	client      *ethclient.Client
	funder      common.Address
	pk          *ecdsa.PrivateKey
	minScore    float64
	captchaHost string
}

func init() {
	var ok bool
	fundingAmount, ok = new(big.Int).SetString(fundingAmountWei, 10)
	if !ok {
		log.Fatal("could not set funding amount")
	}
}

// New faucet server implementation.
func New(
	r recaptcha.Recaptcha,
	rpcPath,
	funderPrivateKey string,
	minScore float64,
	captchaHost string,
) (*Server, error) {
	client, err := ethclient.DialContext(context.Background(), rpcPath)
	if err != nil {
		return nil, fmt.Errorf("could not dial %s: %w", rpcPath, err)
	}

	pk, err := crypto.HexToECDSA(funderPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("could parse funder private key: %w", err)
	}

	funder := crypto.PubkeyToAddress(pk.PublicKey)
	bal, err := client.BalanceAt(context.Background(), funder, nil)
	if err != nil {
		return nil, fmt.Errorf("could retrieve funder's current balance: %w", err)
	}

	log.WithFields(logrus.Fields{
		"fundsInWei": bal,
		"publicKey":  funder.Hex(),
	}).Info("Funder details")
	return &Server{
		r:           r,
		client:      client,
		funder:      funder,
		pk:          pk,
		minScore:    minScore,
		captchaHost: captchaHost,
	}, nil
}

// RequestFunds from the ethereum 1.x faucet. Requires a valid captcha response.
func (s *Server) RequestFunds(
	ctx context.Context, req *faucetpb.FundingRequest,
) (*faucetpb.FundingResponse, error) {
	ipAddress, err := s.getIPAddress(ctx)
	if err != nil {
		log.WithError(err).Error("Could not fetch IP from request")
		return nil, status.Errorf(codes.FailedPrecondition, "Could not get IP address from request: %v", err)
	}

	if err := s.verifyRecaptcha(ipAddress, req); err != nil {
		log.WithError(err).Error("Failed captcha verification")
		return nil, status.Errorf(codes.PermissionDenied, "Failed captcha verification: %v", err)
	}

	// Check if funded too recently and keep track of funded address.
	fundingLock.Lock()
	exceedPeerLimit := ipCounter[ipAddress] >= ipLimit
	if funded[req.WalletAddress] || exceedPeerLimit {
		if exceedPeerLimit {
			log.WithField("ipAddress", ipAddress).Warn("IP trying to get funding despite over request limit")
		}
		fundingLock.Unlock()
		return nil, status.Error(codes.PermissionDenied, "Funded too recently")
	}
	funded[req.WalletAddress] = true
	fundingLock.Unlock()

	txHash, err := s.fundAndWait(common.HexToAddress(req.WalletAddress))
	if err != nil {
		log.WithError(err).Error("Could not send goerli transaction")
		return nil, status.Errorf(codes.Internal, "Could not send goerli transaction: %v", err)
	}
	fundingLock.Lock()
	ipCounter[ipAddress]++
	fundingLock.Unlock()

	log.WithFields(logrus.Fields{
		"txHash":           txHash,
		"requesterAddress": req.WalletAddress,
	}).Info("Funded successfully")
	return &faucetpb.FundingResponse{
		Amount:          fundingAmount.String(),
		TransactionHash: txHash,
	}, nil
}

func (s *Server) fundAndWait(to common.Address) (string, error) {
	nonce := uint64(0)
	nonce, err := s.client.PendingNonceAt(context.Background(), s.funder)
	if err != nil {
		return "", fmt.Errorf("could not get nonce: %w", err)
	}

	tx := types.NewTransaction(
		nonce,
		to,
		fundingAmount,
		txGasLimit,
		big.NewInt(1*params.GWei),
		nil, /*data*/
	)
	tx, err = types.SignTx(tx, types.NewEIP155Signer(big.NewInt(5)), s.pk)
	if err != nil {
		return "", fmt.Errorf("could not sign tx: %w", err)
	}

	if err := s.client.SendTransaction(context.Background(), tx); err != nil {
		return "", fmt.Errorf("could not send tx: %w", err)
	}

	// Wait for transaction to mine.
	for pending := true; pending; _, pending, err = s.client.TransactionByHash(context.Background(), tx.Hash()) {
		if err != nil {
			return "", fmt.Errorf("could not wait for tx to mine: %w", err)
		}
		time.Sleep(1 * time.Second)
	}
	return tx.Hash().Hex(), nil
}

func (s *Server) getIPAddress(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("x-forwarded-for")) < 1 {
		return "", errors.New("metadata not ok")
	}
	address := md.Get("x-forwarded-for")[0]
	return address, nil
}

func (s *Server) verifyRecaptcha(ipAddress string, req *faucetpb.FundingRequest) error {
	log.WithField("ip-address", ipAddress).Info("Verifying captcha...")
	rr, err := s.r.Check(ipAddress, req.CaptchaResponse)
	if err != nil {
		return fmt.Errorf("could not check response: %w", err)
	}
	if !rr.Success {
		return fmt.Errorf("unsuccessful captcha request, error codes: %+v", rr.ErrorCodes)
	}
	if rr.Score < s.minScore {
		return fmt.Errorf("recaptcha score too low (%f)", rr.Score)
	}
	if time.Now().After(rr.ChallengeTS.Add(2 * time.Minute)) {
		return errors.New("captcha challenge too old")
	}
	if rr.Action != req.WalletAddress {
		return fmt.Errorf("action was %s, wanted %s", rr.Action, req.WalletAddress)
	}
	if !strings.HasSuffix(rr.Hostname, s.captchaHost) {
		return fmt.Errorf("expected hostname (%s) to end in %s", rr.Hostname, s.captchaHost)
	}
	return nil
}

// Reduce the counter for each ip every few hours.
func ipAddressCounterWatcher() {
	ticker := time.NewTicker(pruneDuration)
	for {
		<-ticker.C
		fundingLock.Lock()
		for ip, ctr := range ipCounter {
			if ctr == 0 {
				continue
			}
			ipCounter[ip] = ctr - 1
		}
		fundingLock.Unlock()
	}
}
