//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/myy-chat/backend/golang/app/memory/internal/biz"
	"github.com/myy-chat/backend/golang/app/memory/internal/conf"
	"github.com/myy-chat/backend/golang/app/memory/internal/data"
	"github.com/myy-chat/backend/golang/app/memory/internal/server"
	"github.com/myy-chat/backend/golang/app/memory/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
