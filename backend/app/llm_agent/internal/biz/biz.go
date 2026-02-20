package biz

import (
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewUsecase)

// Usecase is the llm_agent business usecase.
type Usecase struct {
	repo Repo
}

// Repo is the llm_agent repository interface.
type Repo interface {
	// TODO: Define your repository methods
}

// NewUsecase creates a new llm_agent usecase.
func NewUsecase(repo Repo) *Usecase {
	return &Usecase{repo: repo}
}
