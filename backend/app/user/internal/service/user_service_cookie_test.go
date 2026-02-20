package service

import (
	"net/http"
	"testing"
	"time"
)

func TestNewRefreshTokenCookie(t *testing.T) {
	cookie := newRefreshTokenCookie("token-value")

	if cookie.Name != refreshTokenCookieName {
		t.Fatalf("cookie name mismatch: got %s", cookie.Name)
	}
	if cookie.Value != "token-value" {
		t.Fatalf("cookie value mismatch: got %s", cookie.Value)
	}
	if cookie.Path != refreshTokenCookiePath {
		t.Fatalf("cookie path mismatch: got %s want %s", cookie.Path, refreshTokenCookiePath)
	}
	if !cookie.HttpOnly {
		t.Fatal("cookie should be HttpOnly")
	}
	if !cookie.Secure {
		t.Fatal("cookie should be Secure")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookie SameSite mismatch: got %v", cookie.SameSite)
	}
	if cookie.MaxAge != refreshTokenCookieMaxAgeSec {
		t.Fatalf("cookie MaxAge mismatch: got %d want %d", cookie.MaxAge, refreshTokenCookieMaxAgeSec)
	}
}

func TestNewClearRefreshTokenCookie(t *testing.T) {
	cookie := newClearRefreshTokenCookie()

	if cookie.Name != refreshTokenCookieName {
		t.Fatalf("cookie name mismatch: got %s", cookie.Name)
	}
	if cookie.Path != refreshTokenCookiePath {
		t.Fatalf("cookie path mismatch: got %s want %s", cookie.Path, refreshTokenCookiePath)
	}
	if cookie.Value != "" {
		t.Fatalf("clear cookie should have empty value, got %q", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Fatalf("clear cookie MaxAge mismatch: got %d", cookie.MaxAge)
	}
	if !cookie.Expires.Equal(time.Unix(0, 0)) {
		t.Fatalf("clear cookie expires mismatch: got %v", cookie.Expires)
	}
	if !cookie.HttpOnly {
		t.Fatal("clear cookie should be HttpOnly")
	}
	if !cookie.Secure {
		t.Fatal("clear cookie should be Secure")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("clear cookie SameSite mismatch: got %v", cookie.SameSite)
	}
}
