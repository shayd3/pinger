package checker

import (
	"fmt"
	"time"
)

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusTimeout   Status = "timeout"
	StatusError     Status = "error"
)

// LatencyMs is a wrapper for time.Duration that serializes to milliseconds
type LatencyMs time.Duration

func (l LatencyMs) MarshalJSON() ([]byte, error) {
	ms := time.Duration(l).Milliseconds()
	return []byte(fmt.Sprintf("%d", ms)), nil
}

type Result struct {
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Status     Status    `json:"status"`
	StatusCode int       `json:"status_code,omitempty"`
	Latency    LatencyMs `json:"latency_ms"`
	Error      string    `json:"error,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

func (r *Result) IsHealthy() bool {
	return r.Status == StatusHealthy
}

func (r *Result) LatencyDuration() time.Duration {
	return time.Duration(r.Latency)
}
