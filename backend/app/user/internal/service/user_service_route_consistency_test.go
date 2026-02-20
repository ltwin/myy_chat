package service

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRefreshCookiePathConsistentWithRoutes(t *testing.T) {
	const expectedRefreshPath = "/api/v1/users/refresh"
	if refreshTokenCookiePath != expectedRefreshPath {
		t.Fatalf("unexpected refresh cookie path: got %s want %s", refreshTokenCookiePath, expectedRefreshPath)
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve runtime caller")
	}

	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../../../"))
	protoPath := filepath.Join(backendRoot, "api/user/v1/user.proto")
	apisixPath := filepath.Clean(filepath.Join(backendRoot, "../deployments/apisix/apisix.yaml"))

	protoContent, err := os.ReadFile(protoPath)
	if err != nil {
		t.Fatalf("read proto failed: %v", err)
	}
	if !strings.Contains(string(protoContent), `post: "/api/v1/users/refresh"`) {
		t.Fatalf("proto refresh route missing expected path %s", expectedRefreshPath)
	}

	apisixContent, err := os.ReadFile(apisixPath)
	if err != nil {
		t.Fatalf("read apisix config failed: %v", err)
	}
	if !strings.Contains(string(apisixContent), "uri: /api/v1/users/refresh") {
		t.Fatalf("apisix refresh route missing expected path %s", expectedRefreshPath)
	}
}
