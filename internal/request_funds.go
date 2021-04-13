package internal

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	faucetpb "github.com/rauljordan/eth-faucet/proto/faucet"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const weiPerETH = 1e18

// RequestFunds from an Ethereum faucet. Requires a valid captcha response.
func (s *Server) RequestFunds(
	ctx context.Context, req *faucetpb.FundingRequest,
) (*faucetpb.FundingResponse, error) {
	if req.WalletAddress == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Request needs a valid ETH wallet address")
	}
	ipAddress, err := s.getIPAddress(ctx)
	if err != nil {
		log.WithError(err).Error("Could not fetch IP from request")
		return nil, status.Errorf(codes.FailedPrecondition, "Could not get IP address from request: %v", err)
	}

	// Verify the provided captcha in the request.
	log.WithField("ipAddress", ipAddress).Info("Verifying captcha...")
	if err := s.verifyRecaptcha(ipAddress, req); err != nil {
		log.WithError(err).Error("Failed captcha verification")
		return nil, status.Errorf(codes.PermissionDenied, "Failed captcha verification: %v", err)
	}

	// Check if ip should be rate limited.
	if !s.rateLimiter.shouldAllowRequest(ipAddress, req.WalletAddress) {
		return nil, status.Error(codes.PermissionDenied, "Funded too recently")
	}

	log.WithFields(logrus.Fields{
		"ipAddress": ipAddress,
		"address":   req.WalletAddress,
	}).Info("Attempting to fund address")
	txHash, err := s.fundAndWait(common.HexToAddress(req.WalletAddress))
	if err != nil {
		log.WithError(err).Error("Could not send goerli transaction")
		return nil, status.Errorf(codes.Internal, "Could not send goerli transaction: %v", err)
	}

	// Mark the ip and Ethereum address pair as funded for the rate limiter.
	s.rateLimiter.markAsFunded(ipAddress, req.WalletAddress)

	log.WithFields(logrus.Fields{
		"txHash":           txHash,
		"requesterAddress": req.WalletAddress,
	}).Info("Funded successfully")

	fundingAmountWei := new(big.Float).SetInt(s.fundingAmount)
	fundedETH := new(big.Float).Quo(fundingAmountWei, big.NewFloat(weiPerETH))
	return &faucetpb.FundingResponse{
		Amount:          fundedETH.String(),
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
		s.fundingAmount,
		s.cfg.GasLimit,
		big.NewInt(1*params.GWei), /* testnet gas price */
		nil,                       /* data */
	)
	tx, err = types.SignTx(tx, types.NewEIP155Signer(big.NewInt(s.cfg.ChainId)), s.pk)
	if err != nil {
		return "", fmt.Errorf("could not sign tx: %w", err)
	}

	if err := s.client.SendTransaction(context.Background(), tx); err != nil {
		return "", fmt.Errorf("could not send tx: %w", err)
	}

	// Wait for transaction to mine.
	log.WithField("txHash", fmt.Sprintf("%#x", tx.Hash())).Info("Awaiting for tx to mine...")
	start := time.Now()
	for pending := true; pending; _, pending, err = s.client.TransactionByHash(context.Background(), tx.Hash()) {
		if err != nil {
			return "", fmt.Errorf("could not wait for tx to mine: %w", err)
		}
		time.Sleep(1 * time.Second)
	}
	log.WithFields(logrus.Fields{
		"timeElapsed": fmt.Sprintf("%v", time.Since(start)),
		"txHash":      fmt.Sprintf("%#x", tx.Hash()),
	}).Info("Transaction mined")
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
