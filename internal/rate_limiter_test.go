package internal

import "testing"

func Test_simpleRateLimiter(t *testing.T) {
	ipLimitPerAddress := 3
	rl := newSimpleRateLimiter(ipLimitPerAddress)
	ethAddress := "0x0101"
	fakeIP := "192.0.0.1"

	t.Run("first_time_request_should_allow", func(t *testing.T) {
		if ok := rl.shouldAllowRequest(fakeIP, ethAddress); !ok {
			t.Error("First time making request should always be allowed")
		}
	})

	t.Run("funded_but_under_ip_rate_limit_disallow", func(t *testing.T) {
		rl.fundedAddresses[ethAddress] = true
		if ok := rl.shouldAllowRequest(fakeIP, ethAddress); ok {
			t.Error("Should disallow after marked as funded")
		}
		rl.fundedAddresses[ethAddress] = false
	})

	t.Run("over_ip_rate_limit_disallow", func(t *testing.T) {
		for i := 0; i < ipLimitPerAddress; i++ {
			rl.markAsFunded(fakeIP, ethAddress)
		}
		if ok := rl.shouldAllowRequest(fakeIP, ethAddress); ok {
			t.Error("Should disallow after reaching rate limit")
		}
	})

	t.Run("reset_rate_limit_should_allow", func(t *testing.T) {
		// Reset the limit.
		rl.fundedAddresses[ethAddress] = false
		rl.ipCounter[fakeIP] = 0

		if ok := rl.shouldAllowRequest(fakeIP, ethAddress); !ok {
			t.Error("Should allow after resetting the limits")
		}
	})
}
