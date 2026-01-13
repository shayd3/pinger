package checker

import (
	"context"

	"github.com/shayd3/pinger/internal/config"
)

type Checker interface {
	Check(ctx context.Context, target config.Target) Result
}

func New(checkType config.CheckType) Checker {
	switch checkType {
	case config.CheckTypeTCP:
		return &TCPChecker{}
	case config.CheckTypeDNS:
		return &DNSChecker{}
	default:
		return &HTTPChecker{}
	}
}
