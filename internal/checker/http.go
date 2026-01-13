package checker

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/shayd3/pinger/internal/config"
)

type HTTPChecker struct{}

func (h *HTTPChecker) Check(ctx context.Context, target config.Target) Result {
	result := Result{
		Name:      target.Name,
		URL:       target.URL,
		Timestamp: time.Now(),
	}

	client := &http.Client{
		Timeout: target.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
	if err != nil {
		result.Status = StatusError
		result.Error = err.Error()
		return result
	}

	for key, value := range target.Headers {
		req.Header.Set(key, value)
	}

	start := time.Now()
	resp, err := client.Do(req)
	result.Latency = LatencyMs(time.Since(start))

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = StatusTimeout
			result.Error = "request timed out"
		} else {
			result.Status = StatusError
			result.Error = err.Error()
		}
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	if resp.StatusCode == target.ExpectedStatus {
		result.Status = StatusHealthy
	} else {
		result.Status = StatusUnhealthy
		result.Error = "unexpected status code"
	}

	return result
}
