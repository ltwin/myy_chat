package service

import (
	"github.com/google/wire"

	"github.com/myy-chat/backend/golang/app/conversation/internal/biz"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewService)

// Service is the conversation service implementation.
type Service struct {
	uc *biz.Usecase
}

// NewService creates a new conversation service.
func NewService(uc *biz.Usecase) *Service {
	return &Service{uc: uc}
}
