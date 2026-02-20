package server

import (
	"context"
	"errors"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	kratosmiddleware "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"

	v1 "github.com/myy-chat/backend/api/user/v1"
	"github.com/myy-chat/backend/app/user/internal/security"
	authmiddleware "github.com/myy-chat/backend/pkg/middleware"
	redisclient "github.com/myy-chat/backend/pkg/redis"
)

func jwtSelectorMiddleware(redis *redisclient.Client) kratosmiddleware.Middleware {
	jwtConfig := authmiddleware.DefaultJWTConfig([]byte(security.ResolveJWTSigningKey()))
	jwtConfig.ErrorHandler = func(ctx context.Context, err error) error {
		switch {
		case errors.Is(err, authmiddleware.ErrBlacklistCheck):
			return kerrors.ServiceUnavailable("AUTH_BLACKLIST_ERROR", "failed to verify token revocation status")
		case errors.Is(err, authmiddleware.ErrMissingToken),
			errors.Is(err, authmiddleware.ErrInvalidToken),
			errors.Is(err, authmiddleware.ErrInvalidClaims),
			errors.Is(err, authmiddleware.ErrExpiredToken),
			errors.Is(err, authmiddleware.ErrInvalidSignature),
			errors.Is(err, authmiddleware.ErrTokenRevoked):
			return kerrors.Unauthorized("UNAUTHORIZED", "authentication required")
		default:
			return kerrors.Unauthorized("UNAUTHORIZED", "authentication required")
		}
	}

	if redis != nil {
		jwtConfig.BlacklistChecker = authmiddleware.NewRedisTokenBlacklist(redis)
	}

	return selector.Server(authmiddleware.JWTAuth(jwtConfig)).
		Path(
			v1.OperationUserServiceGetUser,
			v1.OperationUserServiceUpdateUser,
			v1.OperationUserServiceUpdatePassword,
			v1.OperationUserServiceGetProfile,
			v1.OperationUserServiceUpdateProfile,
			v1.OperationUserServiceRequestAccountDeletion,
			v1.OperationUserServiceCancelAccountDeletion,
			v1.OperationUserServiceLogout,
			v1.OperationUserServiceLogoutAll,
		).
		Build()
}

// NewAuthRedisClient creates one shared Redis client for auth-related components.
func NewAuthRedisClient(logger log.Logger) (*redisclient.Client, func(), error) {
	cfg := redisclient.DefaultConfig()
	cfg.Addr = security.ResolveRedisAddr()

	client, err := redisclient.NewClient(cfg)
	if err != nil {
		if security.IsProduction() {
			return nil, nil, err
		}
		log.NewHelper(logger).Warnf("redis disabled for auth components: %v", err)
		return nil, func() {}, nil
	}

	cleanup := func() {
		_ = client.Close()
	}
	return client, cleanup, nil
}
