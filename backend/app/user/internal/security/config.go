package security

import (
	"os"
	"strings"
)

const (
	defaultJWTSigningKey = "myy-chat-jwt-secret-key-for-testing"
	defaultRedisAddr     = "localhost:6379"
)

// IsProduction returns true when runtime environment is production-like.
func IsProduction() bool {
	for _, key := range []string{"APP_ENV", "GO_ENV", "ENV"} {
		value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
		if value == "production" || value == "prod" {
			return true
		}
	}
	return false
}

func ResolveJWTSigningKey() string {
	for _, key := range []string{"MYY_CHAT_JWT_SIGNING_KEY", "JWT_SIGNING_KEY"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	if IsProduction() {
		panic("missing JWT signing key: set MYY_CHAT_JWT_SIGNING_KEY or JWT_SIGNING_KEY")
	}
	return defaultJWTSigningKey
}

func ResolveRedisAddr() string {
	for _, key := range []string{"MYY_CHAT_REDIS_ADDR", "REDIS_ADDR"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	if IsProduction() {
		panic("missing Redis addr: set MYY_CHAT_REDIS_ADDR or REDIS_ADDR")
	}
	return defaultRedisAddr
}

func LegacyRefreshTokenInBodyEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("MYY_CHAT_LEGACY_REFRESH_TOKEN_IN_BODY")))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
