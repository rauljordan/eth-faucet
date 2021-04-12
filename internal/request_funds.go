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
