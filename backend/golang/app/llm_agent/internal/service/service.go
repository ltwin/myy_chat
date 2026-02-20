package service

import (
	"github.com/google/wire"

	"github.com/myy-chat/backend/golang/app/llm_agent/internal/biz"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewService)

// Service is the llm_agent service implementation.
type Service struct {
	uc *biz.Usecase
}

// NewService creates a new llm_agent service.
func NewService(uc *biz.Usecase) *Service {
	return &Service{uc: uc}
}
