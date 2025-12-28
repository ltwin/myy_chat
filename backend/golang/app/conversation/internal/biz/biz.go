package biz

import (
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewUsecase)

// Usecase is the conversation business usecase.
type Usecase struct {
	repo Repo
}

// Repo is the conversation repository interface.
type Repo interface {
	// TODO: Define your repository methods
}

// NewUsecase creates a new conversation usecase.
func NewUsecase(repo Repo) *Usecase {
	return &Usecase{repo: repo}
}
