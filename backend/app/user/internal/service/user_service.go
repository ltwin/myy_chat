// Package service gRPC 服务层
package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	pb "github.com/myy-chat/backend/api/user/v1"
	"github.com/myy-chat/backend/app/user/internal/biz"
	"github.com/myy-chat/backend/pkg/middleware"
	redisclient "github.com/myy-chat/backend/pkg/redis"
)

// UserService gRPC 用户服务实现
type UserService struct {
	pb.UnimplementedUserServiceServer

	userService  *biz.UserService
	jwtGenerator *middleware.JWTGenerator
	blacklist    *middleware.RedisTokenBlacklist
}

const (
	refreshTokenCookieName      = "rt"
	refreshTokenCookiePath      = "/api/v1/users/refresh"
	refreshTokenCookieMaxAgeSec = 7 * 24 * 60 * 60
	accessTokenExpiresInSec     = 15 * 60
	defaultJWTSigningKey        = "myy-chat-jwt-secret-key-for-testing"
	defaultRedisAddr            = "localhost:6379"
)

// NewUserService 创建 gRPC 用户服务
func NewUserService(us *biz.UserService) *UserService {
	jwtConfig := middleware.DefaultJWTConfig([]byte(resolveJWTSigningKey()))
	return &UserService{
		userService:  us,
		jwtGenerator: middleware.NewJWTGenerator(jwtConfig),
		blacklist:    buildTokenBlacklist(),
	}
}

// Register 用户注册 (FR-001)
func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	output, err := s.userService.Register(ctx, biz.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	sessionOutput, err := s.userService.IssueSession(ctx, output.User.ID, extractUserAgent(ctx), extractClientIP(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}
	accessToken, err := s.jwtGenerator.GenerateTokenWithSession(output.User.ID, output.User.Username, sessionOutput.Session.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}
	setRefreshTokenCookie(ctx, sessionOutput.RefreshToken)

	return &pb.RegisterResponse{
		UserId:                    output.User.ID,
		AccessToken:               accessToken,
		RefreshToken:              sessionOutput.RefreshToken, // deprecated: 兼容旧客户端
		ExpiresIn:                 accessTokenExpiresInSec,
		EmailVerificationRequired: !output.User.EmailVerified,
	}, nil
}

// Login 用户登录 (FR-003)
func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	output, err := s.userService.Login(ctx, biz.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	sessionOutput, err := s.userService.IssueSession(ctx, output.User.ID, extractUserAgent(ctx), extractClientIP(ctx))
	if err != nil {
		return nil, toGRPCError(err)
	}
	accessToken, err := s.jwtGenerator.GenerateTokenWithSession(output.User.ID, output.User.Username, sessionOutput.Session.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}
	setRefreshTokenCookie(ctx, sessionOutput.RefreshToken)

	return &pb.LoginResponse{
		UserId:       output.User.ID,
		AccessToken:  accessToken,
		RefreshToken: sessionOutput.RefreshToken, // deprecated: 兼容旧客户端
		ExpiresIn:    accessTokenExpiresInSec,
		User:         userToProto(output.User),
	}, nil
}

// Logout 用户登出
func (s *UserService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	claims, err := extractClaimsFromContextOrAuthorization(ctx, s.jwtGenerator)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	if err := s.userService.RevokeSession(ctx, claims.SessionID); err != nil {
		return nil, toGRPCError(err)
	}
	if err := s.revokeTokenImmediately(ctx, claims); err != nil {
		return nil, status.Error(codes.Internal, "failed to persist token revocation")
	}

	clearRefreshTokenCookie(ctx)
	return &pb.LogoutResponse{
		Success: true,
	}, nil
}

// LogoutAll 用户退出所有设备
func (s *UserService) LogoutAll(ctx context.Context, req *pb.LogoutAllRequest) (*pb.LogoutAllResponse, error) {
	claims, _ := extractClaimsFromContextOrAuthorization(ctx, s.jwtGenerator)

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		if claims == nil {
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}
		userID = claims.UserID
	}
	sessions, err := s.userService.RevokeAllSessions(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	if err := s.revokeAllSessionsImmediately(ctx, sessions); err != nil {
		return nil, status.Error(codes.Internal, "failed to persist token revocation")
	}
	if claims != nil {
		if err := s.revokeTokenImmediately(ctx, claims); err != nil {
			return nil, status.Error(codes.Internal, "failed to persist token revocation")
		}
	}

	clearRefreshTokenCookie(ctx)
	return &pb.LogoutAllResponse{
		Success: true,
	}, nil
}

// RefreshToken 刷新令牌
func (s *UserService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	refreshToken, err := extractRefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "refresh token required")
	}

	output, err := s.userService.RefreshSession(ctx, refreshToken)
	if err != nil {
		return nil, toGRPCError(err)
	}
	accessToken, err := s.jwtGenerator.GenerateTokenWithSession(output.User.ID, output.User.Username, output.Session.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}
	setRefreshTokenCookie(ctx, output.RefreshToken)

	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: output.RefreshToken, // deprecated: 兼容旧客户端
		ExpiresIn:    accessTokenExpiresInSec,
	}, nil
}

// GetUser 获取当前用户信息
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	// 获取积分余额
	balance, _ := s.userService.GetCreditBalance(ctx, userID)

	protoUser := userToProto(user)
	protoUser.CreditBalance = balance

	return &pb.GetUserResponse{
		User: protoUser,
	}, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	input := biz.UpdateUserInput{
		UserID: userID,
	}
	if req.Username != "" {
		input.Username = &req.Username
	}
	if req.Phone != "" {
		input.Phone = &req.Phone
	}
	if req.AvatarUrl != "" {
		input.AvatarURL = &req.AvatarUrl
	}

	user, err := s.userService.UpdateUser(ctx, input)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UpdateUserResponse{
		User: userToProto(user),
	}, nil
}

// UpdatePassword 修改密码 (FR-003c)
func (s *UserService) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.userService.UpdatePassword(ctx, userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UpdatePasswordResponse{
		Success: true,
	}, nil
}

// GetProfile 获取用户画像
func (s *UserService) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := s.userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetProfileResponse{
		Profile: profileToProto(profile),
	}, nil
}

// UpdateProfile 更新用户画像
func (s *UserService) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var location *biz.Location
	if req.Location != nil {
		location = &biz.Location{
			Country:  req.Location.Country,
			Province: req.Location.Province,
			City:     req.Location.City,
		}
	}

	profile, err := s.userService.UpdateProfile(ctx, biz.UpdateProfileInput{
		UserID:     userID,
		FullName:   req.FullName,
		Gender:     req.Gender,
		BirthDate:  req.BirthDate,
		Location:   location,
		Interests:  req.Interests,
		Occupation: req.Occupation,
		Bio:        req.Bio,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UpdateProfileResponse{
		Profile: profileToProto(profile),
	}, nil
}

// RequestAccountDeletion 请求账号删除 (FR-004a)
func (s *UserService) RequestAccountDeletion(ctx context.Context, req *pb.RequestAccountDeletionRequest) (*pb.RequestAccountDeletionResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	deletionTime, err := s.userService.RequestAccountDeletion(ctx, userID, req.Password)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.RequestAccountDeletionResponse{
		Success:             true,
		DeletionScheduledAt: timestamppb.New(deletionTime),
	}, nil
}

// CancelAccountDeletion 取消账号删除 (FR-004b)
func (s *UserService) CancelAccountDeletion(ctx context.Context, req *pb.CancelAccountDeletionRequest) (*pb.CancelAccountDeletionResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.userService.CancelAccountDeletion(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CancelAccountDeletionResponse{
		Success: true,
	}, nil
}

// VerifyEmail 验证邮箱 (FR-002)
func (s *UserService) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	// TODO: 从 token 中获取用户 ID 并验证
	return &pb.VerifyEmailResponse{
		Success: true,
	}, nil
}

// ResendVerification 重发验证邮件
func (s *UserService) ResendVerification(ctx context.Context, req *pb.ResendVerificationRequest) (*pb.ResendVerificationResponse, error) {
	// TODO: 实现邮件发送逻辑
	return &pb.ResendVerificationResponse{
		Success: true,
		Message: "Verification email sent",
	}, nil
}

// ForgotPassword 忘记密码
func (s *UserService) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	// TODO: 实现密码重置邮件发送
	return &pb.ForgotPasswordResponse{
		Success: true,
		Message: "Password reset email sent",
	}, nil
}

// ResetPassword 重置密码
func (s *UserService) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	// TODO: 验证 token 并重置密码
	return &pb.ResetPasswordResponse{
		Success: true,
	}, nil
}

// 辅助函数

func resolveJWTSigningKey() string {
	for _, envKey := range []string{"MYY_CHAT_JWT_SIGNING_KEY", "JWT_SIGNING_KEY"} {
		if signingKey := strings.TrimSpace(os.Getenv(envKey)); signingKey != "" {
			return signingKey
		}
	}
	return defaultJWTSigningKey
}

func buildTokenBlacklist() *middleware.RedisTokenBlacklist {
	cfg := redisclient.DefaultConfig()
	cfg.Addr = defaultRedisAddr
	if redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDR")); redisAddr != "" {
		cfg.Addr = redisAddr
	}

	client, err := redisclient.NewClient(cfg)
	if err != nil {
		return nil
	}
	return middleware.NewRedisTokenBlacklist(client)
}

func extractRefreshToken(ctx context.Context, fallback string) (string, error) {
	if req, ok := khttp.RequestFromServerContext(ctx); ok {
		if cookie, err := req.Cookie(refreshTokenCookieName); err == nil {
			if token := strings.TrimSpace(cookie.Value); token != "" {
				return token, nil
			}
		}
	}

	token := strings.TrimSpace(fallback)
	if token != "" {
		return token, nil
	}
	return "", errors.New("refresh token required")
}

func extractClaimsFromContextOrAuthorization(ctx context.Context, generator *middleware.JWTGenerator) (*middleware.UserClaims, error) {
	if claims, ok := middleware.ClaimsFromContext(ctx); ok {
		if userClaims, ok := claims.(*middleware.UserClaims); ok {
			return userClaims, nil
		}
	}

	token := extractBearerTokenFromContext(ctx)
	if token == "" {
		return nil, middleware.ErrMissingToken
	}
	return generator.ParseToken(token)
}

func extractBearerTokenFromContext(ctx context.Context) string {
	req, ok := khttp.RequestFromServerContext(ctx)
	if !ok {
		return ""
	}

	auth := strings.TrimSpace(req.Header.Get("Authorization"))
	if auth == "" {
		return ""
	}

	prefix := "Bearer "
	if strings.HasPrefix(strings.ToLower(auth), strings.ToLower(prefix)) {
		return strings.TrimSpace(auth[len(prefix):])
	}
	return auth
}

func extractUserAgent(ctx context.Context) string {
	req, ok := khttp.RequestFromServerContext(ctx)
	if !ok {
		return ""
	}
	return req.UserAgent()
}

func extractClientIP(ctx context.Context) string {
	req, ok := khttp.RequestFromServerContext(ctx)
	if !ok {
		return ""
	}

	if forwardedFor := strings.TrimSpace(req.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := strings.TrimSpace(req.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		return host
	}
	return strings.TrimSpace(req.RemoteAddr)
}

func (s *UserService) revokeTokenImmediately(ctx context.Context, claims *middleware.UserClaims) error {
	if s.blacklist == nil || claims == nil {
		return nil
	}

	if err := s.blacklist.RevokeSID(ctx, claims.SessionID, 7*24*time.Hour); err != nil {
		return err
	}

	jtiTTL := 15 * time.Minute
	if claims.ExpiresAt != nil {
		if ttl := time.Until(claims.ExpiresAt.Time); ttl > 0 {
			jtiTTL = ttl
		}
	}
	return s.blacklist.RevokeJTI(ctx, claims.ID, jtiTTL)
}

func (s *UserService) revokeAllSessionsImmediately(ctx context.Context, sessions []*biz.Session) error {
	if s.blacklist == nil {
		return nil
	}

	for _, session := range sessions {
		if session == nil || session.ID <= 0 {
			continue
		}
		if err := s.blacklist.RevokeSID(ctx, session.ID, 7*24*time.Hour); err != nil {
			return err
		}
	}
	return nil
}

// getUserIDFromContext 从上下文获取用户ID
func getUserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "authentication required")
	}
	if userID <= 0 {
		return 0, status.Error(codes.InvalidArgument, "invalid user id")
	}
	return userID, nil
}

// toGRPCError 将业务错误转换为 gRPC 状态错误
func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	// 已经是 gRPC status error
	if _, ok := status.FromError(err); ok {
		return err
	}

	// 根据错误类型转换
	switch {
	case errors.Is(err, biz.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, biz.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, biz.ErrInvalidPassword):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, biz.ErrAccountLocked):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, biz.ErrEmailNotVerified):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, biz.ErrInvalidRefreshToken), errors.Is(err, biz.ErrSessionExpired), errors.Is(err, biz.ErrSessionRevoked):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

// userToProto 将用户实体转换为 protobuf 消息
func userToProto(user *biz.User) *pb.User {
	if user == nil {
		return nil
	}

	protoUser := &pb.User{
		Id:                user.ID,
		Username:          user.Username,
		Email:             user.Email,
		Phone:             user.Phone,
		AvatarUrl:         user.AvatarURL,
		EmailVerified:     user.EmailVerified,
		CreatedAt:         timestamppb.New(user.CreatedAt),
		DeletionScheduled: user.IsDeletionScheduled(),
	}

	if user.LastLoginAt != nil {
		protoUser.LastLoginAt = timestamppb.New(*user.LastLoginAt)
	}

	return protoUser
}

// profileToProto 将用户画像转换为 protobuf 消息
func profileToProto(profile *biz.UserProfile) *pb.UserProfile {
	if profile == nil {
		return nil
	}

	protoProfile := &pb.UserProfile{
		UserId:     profile.UserID,
		FullName:   profile.FullName,
		Gender:     profile.Gender,
		Interests:  profile.Interests,
		Occupation: profile.Occupation,
		Bio:        profile.Bio,
		UpdatedAt:  timestamppb.New(profile.UpdatedAt),
	}

	if profile.BirthDate != nil {
		protoProfile.BirthDate = profile.BirthDate.String()
	}

	if profile.Location != nil {
		protoProfile.Location = &pb.Location{
			Country:  profile.Location.Country,
			Province: profile.Location.Province,
			City:     profile.Location.City,
		}
	}

	return protoProfile
}

func setRefreshTokenCookie(ctx context.Context, token string) {
	khttp.SetCookie(ctx, newRefreshTokenCookie(token))
}

func clearRefreshTokenCookie(ctx context.Context) {
	khttp.SetCookie(ctx, newClearRefreshTokenCookie())
}

func newRefreshTokenCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     refreshTokenCookiePath,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   refreshTokenCookieMaxAgeSec,
	}
}

func newClearRefreshTokenCookie() *http.Cookie {
	return &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     refreshTokenCookiePath,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
}
