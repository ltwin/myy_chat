package service

import (
	"github.com/google/wire"

	"github.com/myy-chat/backend/golang/app/demo/internal/biz"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewService)

// Service is the demo service implementation.
type Service struct {
	uc *biz.Usecase
}

// NewService creates a new demo service.
func NewService(uc *biz.Usecase) *Service {
	return &Service{uc: uc}
}
