package service

import (
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewService)

// Service is the {{SERVICE}} service implementation.
type Service struct {
	// TODO: Add your dependencies here
}

// NewService creates a new {{SERVICE}} service.
func NewService() *Service {
	return &Service{}
}
