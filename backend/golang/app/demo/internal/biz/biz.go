package biz

import (
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewUsecase)

// Usecase is the demo business usecase.
type Usecase struct {
	repo Repo
}

// Repo is the demo repository interface.
type Repo interface {
	// TODO: Define your repository methods
}

// NewUsecase creates a new demo usecase.
func NewUsecase(repo Repo) *Usecase {
	return &Usecase{repo: repo}
}
