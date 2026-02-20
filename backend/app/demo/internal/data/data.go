package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/myy-chat/backend/golang/app/demo/internal/biz"
	"github.com/myy-chat/backend/golang/app/demo/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewRepo)

// Data is the data layer wrapper.
type Data struct {
	// TODO: Add database clients (PostgreSQL, Redis, etc.)
}

// NewData creates a new data layer.
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{}, cleanup, nil
}

// repo implements biz.Repo interface.
type repo struct {
	data *Data
	log  *log.Helper
}

// NewRepo creates a new repository.
func NewRepo(data *Data, logger log.Logger) biz.Repo {
	return &repo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
