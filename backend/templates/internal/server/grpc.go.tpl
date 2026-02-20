package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	"github.com/myy-chat/backend/golang/app/{{SERVICE}}/internal/conf"
	"github.com/myy-chat/backend/golang/app/{{SERVICE}}/internal/service"
	// TODO: Uncomment after generating proto code
	// v1 "github.com/myy-chat/backend/golang/api/{{SERVICE}}/v1"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, svc *service.Service, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	// TODO: Register your gRPC service after proto generation
	// v1.RegisterServiceServer(srv, svc)
	return srv
}
