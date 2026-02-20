package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/myy-chat/backend/app/memory/internal/conf"
	"github.com/myy-chat/backend/app/memory/internal/service"
	// TODO: Uncomment after generating proto code
	// v1 "github.com/myy-chat/backend/api/memory/v1"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, svc *service.Service, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	// TODO: Register your HTTP service after proto generation
	// v1.RegisterServiceHTTPServer(srv, svc)
	return srv
}
