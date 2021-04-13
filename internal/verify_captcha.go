package internal

import (
	"errors"
	"fmt"
	"strings"
	"time"

	faucetpb "github.com/rauljordan/eth-faucet/proto/faucet"
)

func (s *Server) verifyRecaptcha(ipAddress string, req *faucetpb.FundingRequest) error {
	rr, err := s.captcha.Check(ipAddress, req.CaptchaResponse)
	if err != nil {
		return fmt.Errorf("could not check response: %w", err)
	}
	if !rr.Success {
		return fmt.Errorf("unsuccessful captcha request, error codes: %+v", rr.ErrorCodes)
	}
	if rr.Score < s.cfg.CaptchaMinScore {
		return fmt.Errorf("recaptcha score too low (%f)", rr.Score)
	}
	if time.Now().After(rr.ChallengeTS.Add(2 * time.Minute)) {
		return errors.New("captcha challenge too old")
	}
	if rr.Action != req.WalletAddress {
		return fmt.Errorf("action was %s, wanted %s", rr.Action, req.WalletAddress)
	}
	if !strings.HasSuffix(rr.Hostname, s.cfg.CaptchaHost) {
		return fmt.Errorf("expected hostname (%s) to end in %s", rr.Hostname, s.cfg.CaptchaHost)
	}
	return nil
}
