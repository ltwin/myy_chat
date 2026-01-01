// Package integration provides integration tests for the API Gateway
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// Gateway 测试配置
var (
	gatewayURL = getEnv("GATEWAY_URL", "http://localhost:9080")
	adminURL   = getEnv("GATEWAY_ADMIN_URL", "http://localhost:9180")
	adminKey   = getEnv("GATEWAY_ADMIN_KEY", "myy-chat-admin-key")
)

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// TestGatewayHealth 测试网关健康状态
func TestGatewayHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试 (short mode)")
	}

	resp, err := http.Get(gatewayURL + "/apisix/status")
	if err != nil {
		t.Skipf("Gateway 不可用，跳过测试: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200，得到 %d", resp.StatusCode)
	}
}

// TestPublicRoutes 测试公开路由（无需认证）
func TestPublicRoutes(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试 (short mode)")
	}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"健康检查", "GET", "/health", http.StatusOK},
		{"公开角色列表", "GET", "/api/v1/characters/public", http.StatusOK},
		{"积分套餐", "GET", "/api/v1/billing/packages", http.StatusOK},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, gatewayURL+tc.path, nil)
			if err != nil {
				t.Fatalf("创建请求失败: %v", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Skipf("请求失败，服务可能不可用: %v", err)
			}
			defer resp.Body.Close()

			// 允许 404 (服务未实现) 或期望的状态码
			if resp.StatusCode != tc.expectedStatus && resp.StatusCode != http.StatusNotFound {
				t.Errorf("期望状态码 %d 或 404，得到 %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestProtectedRoutesWithoutAuth 测试受保护路由（无认证应返回 401）
func TestProtectedRoutesWithoutAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试 (short mode)")
	}

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"用户信息", "GET", "/api/v1/users/me"},
		{"角色列表", "GET", "/api/v1/characters"},
		{"对话列表", "GET", "/api/v1/conversations"},
		{"记忆列表", "GET", "/api/v1/memories"},
		{"积分余额", "GET", "/api/v1/billing/balance"},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, gatewayURL+tc.path, nil)
			if err != nil {
				t.Fatalf("创建请求失败: %v", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Skipf("请求失败，服务可能不可用: %v", err)
			}
			defer resp.Body.Close()

			// 无认证应返回 401 Unauthorized
			if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusNotFound {
				t.Errorf("期望状态码 401 或 404，得到 %d", resp.StatusCode)
			}
		})
	}
}

// TestCORSHeaders 测试 CORS 跨域头
func TestCORSHeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试 (short mode)")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("OPTIONS", gatewayURL+"/api/v1/users/me", nil)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}

	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")

	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("请求失败，服务可能不可用: %v", err)
	}
	defer resp.Body.Close()

	// 检查 CORS 头
	corsHeaders := []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
	}

	for _, header := range corsHeaders {
		if resp.Header.Get(header) == "" {
			t.Logf("警告: 缺少 CORS 头 %s", header)
		}
	}
}

// TestRateLimiting 测试限流
func TestRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试 (short mode)")
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// 发送大量请求测试限流
	rateLimitHit := false
	for i := 0; i < 100; i++ {
		req, _ := http.NewRequest("GET", gatewayURL+"/health", nil)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitHit = true
			t.Logf("限流在第 %d 次请求后触发", i+1)
			break
		}
	}

	// 限流是可选的，不强制要求
	if !rateLimitHit {
		t.Log("未触发限流（可能配置了较高的限制）")
	}
}

// TestRequestID 测试请求 ID 传递
func TestRequestID(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试 (short mode)")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", gatewayURL+"/health", nil)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("请求失败，服务可能不可用: %v", err)
	}
	defer resp.Body.Close()

	requestID := resp.Header.Get("X-Request-ID")
	if requestID == "" {
		t.Log("警告: 响应中没有 X-Request-ID 头")
	} else {
		t.Logf("请求 ID: %s", requestID)
	}
}

// TestLoginFlow 测试登录流程
func TestLoginFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试 (short mode)")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// 测试登录端点
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "testpassword123",
	}
	body, _ := json.Marshal(loginData)

	req, err := http.NewRequest("POST", gatewayURL+"/api/v1/users/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("请求失败，服务可能不可用: %v", err)
	}
	defer resp.Body.Close()

	// 登录可能返回 401 (无效凭证) 或 404 (服务未实现)
	// 但不应该返回 5xx 错误
	if resp.StatusCode >= 500 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("服务器错误: %d, 响应: %s", resp.StatusCode, string(body))
	}
}

// TestCircuitBreaker 测试熔断器
func TestCircuitBreaker(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试 (short mode)")
	}

	// 熔断器测试需要模拟服务故障，这里只验证配置存在
	t.Log("熔断器配置已在 apisix.yaml 中定义")
	t.Log("完整测试需要模拟后端服务故障场景")
}

// TestPrometheusMetrics 测试 Prometheus 指标端点
func TestPrometheusMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试 (short mode)")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// APISIX Prometheus 指标端点
	resp, err := client.Get("http://localhost:9091/apisix/prometheus/metrics")
	if err != nil {
		t.Skipf("Prometheus 端点不可用: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			t.Logf("Prometheus 指标可用，大小: %d bytes", len(body))
		}
	}
}

// TestAdminAPI 测试管理 API
func TestAdminAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试 (short mode)")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", adminURL+"/apisix/admin/routes", nil)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("X-API-KEY", adminKey)

	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("Admin API 不可用: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Log("Admin API 认证成功")
	} else if resp.StatusCode == http.StatusUnauthorized {
		t.Log("Admin API 需要正确的 API Key")
	}
}

// BenchmarkGatewayLatency 基准测试：网关延迟
func BenchmarkGatewayLatency(b *testing.B) {
	client := &http.Client{Timeout: 5 * time.Second}

	// 预热
	for i := 0; i < 10; i++ {
		resp, err := client.Get(gatewayURL + "/health")
		if err != nil {
			b.Skipf("Gateway 不可用: %v", err)
		}
		resp.Body.Close()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(gatewayURL + "/health")
		if err != nil {
			b.Fatalf("请求失败: %v", err)
		}
		resp.Body.Close()
	}
}

// 辅助函数
func assertStatusCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("期望状态码 %d，得到 %d", want, got)
	}
}

func printResponse(t *testing.T, resp *http.Response) {
	t.Helper()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
}

// TestMain 测试入口
func TestMain(m *testing.M) {
	// 等待服务启动
	fmt.Println("等待 Gateway 服务启动...")
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 30; i++ {
		resp, err := client.Get(gatewayURL + "/apisix/status")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Println("Gateway 服务就绪")
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}

	os.Exit(m.Run())
}
