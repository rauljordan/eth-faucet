package internal

import (
	"context"
	"sync"
	"time"
)

type rateLimiter interface {
	refreshLimits(ctx context.Context)
	shouldAllowRequest(ipAddress, ethAddress string) bool
	markAsFunded(ipAddress, ethAddress string)
}

// Simple rate limiter uses a basic strategy of keeping ip addresses
// in memory and limiting requests to a max limit of ip addresses per
// ETH address requesting faucet funds.
type simpleRateLimiter struct {
	mutex                sync.RWMutex
	ipLimitPerAddress    int
	fundedAddresses      map[string]bool
	ipCounter            map[string]int
	limitRefreshInterval time.Duration
}

func newSimpleRateLimiter(ipLimitPerAddress int) *simpleRateLimiter {
	return &simpleRateLimiter{
		ipLimitPerAddress:    ipLimitPerAddress,
		fundedAddresses:      make(map[string]bool),
		ipCounter:            make(map[string]int),
		limitRefreshInterval: time.Hour * 4, /* Reset limits every 4 hours */
	}
}

func (s *simpleRateLimiter) shouldAllowRequest(ipAddress, ethAddress string) bool {
	s.mutex.RLock()
	exceedPeerLimit := s.ipCounter[ipAddress] >= s.ipLimitPerAddress
	funded := s.fundedAddresses[ethAddress]
	s.mutex.RUnlock()
	if exceedPeerLimit {
		log.WithField(
			"ipAddress", ipAddress,
		).Warn("IP trying to get funding despite over request limit")
	}
	return !(funded || exceedPeerLimit)
}

func (s *simpleRateLimiter) markAsFunded(ipAddress, ethAddress string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ipCounter[ipAddress]++
	s.fundedAddresses[ethAddress] = true
}

// Reduce the counter for each ip every few hours.
func (s *simpleRateLimiter) refreshLimits(ctx context.Context) {
	ticker := time.NewTicker(s.limitRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mutex.Lock()
			log.WithField(
				"numIPs", len(s.ipCounter),
			).Info("Decreasing requests counter for all recorded IP addresses")
			for ip, ctr := range s.ipCounter {
				if ctr == 0 {
					continue
				}
				s.ipCounter[ip] = ctr - 1
			}
			s.mutex.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
