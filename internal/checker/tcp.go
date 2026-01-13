package checker

import (
	"context"
	"net"
	"time"

	"github.com/shayd3/pinger/internal/config"
)

type TCPChecker struct{}

func (t *TCPChecker) Check(ctx context.Context, target config.Target) Result {
	result := Result{
		Name:      target.Name,
		URL:       target.URL,
		Timestamp: time.Now(),
	}

	dialer := &net.Dialer{
		Timeout: target.Timeout,
	}

	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", target.URL)
	result.Latency = LatencyMs(time.Since(start))

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = StatusTimeout
			result.Error = "connection timed out"
		} else {
			result.Status = StatusUnhealthy
			result.Error = err.Error()
		}
		return result
	}
	defer conn.Close()

	result.Status = StatusHealthy
	return result
}
