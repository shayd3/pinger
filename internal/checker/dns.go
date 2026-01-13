package checker

import (
	"context"
	"net"
	"time"

	"github.com/shayd3/pinger/internal/config"
)

type DNSChecker struct{}

func (d *DNSChecker) Check(ctx context.Context, target config.Target) Result {
	result := Result{
		Name:      target.Name,
		URL:       target.URL,
		Timestamp: time.Now(),
	}

	resolver := &net.Resolver{}

	timeoutCtx, cancel := context.WithTimeout(ctx, target.Timeout)
	defer cancel()

	start := time.Now()
	addrs, err := resolver.LookupHost(timeoutCtx, target.URL)
	result.Latency = LatencyMs(time.Since(start))

	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			result.Status = StatusTimeout
			result.Error = "DNS lookup timed out"
		} else {
			result.Status = StatusUnhealthy
			result.Error = err.Error()
		}
		return result
	}

	if len(addrs) == 0 {
		result.Status = StatusUnhealthy
		result.Error = "no addresses found"
		return result
	}

	result.Status = StatusHealthy
	return result
}
