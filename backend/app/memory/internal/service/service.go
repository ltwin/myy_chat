package service

import (
	"github.com/google/wire"

	"github.com/myy-chat/backend/golang/app/memory/internal/biz"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewService)

// Service is the memory service implementation.
type Service struct {
	uc *biz.Usecase
}

// NewService creates a new memory service.
func NewService(uc *biz.Usecase) *Service {
	return &Service{uc: uc}
}
