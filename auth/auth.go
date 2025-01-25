package auth

import (
	"context"
	"errors"
)

var (
	// ErrInvalidToken 无效的令牌
	ErrInvalidToken = errors.New("invalid token")
	// ErrMissingToken 缺少令牌
	ErrMissingToken = errors.New("missing token")
	// ErrExpiredToken 令牌已过期
	ErrExpiredToken = errors.New("token expired")
	// ErrInvalidClaims 无效的声明
	ErrInvalidClaims = errors.New("invalid claims")
)

// Claims 令牌声明
type Claims struct {
	// UserID 用户ID
	UserID string `json:"user_id"`
	// Username 用户名
	Username string `json:"username"`
	// Role 角色
	Role string `json:"role"`
	// ExpiresAt 过期时间
	ExpiresAt int64 `json:"expires_at"`
	// IssuedAt 签发时间
	IssuedAt int64 `json:"issued_at"`
}

// Authenticator 认证器接口
type Authenticator interface {
	// GenerateToken 生成令牌
	GenerateToken(claims Claims) (string, error)
	// ValidateToken 验证令牌
	ValidateToken(token string) (*Claims, error)
}

// ContextKey 上下文键类型
type ContextKey string

const (
	// ClaimsKey 声明上下文键
	ClaimsKey ContextKey = "claims"
)

// FromContext 从上下文中获取声明
func FromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*Claims)
	return claims, ok
}

// NewContext 创建带有声明的上下文
func NewContext(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ClaimsKey, claims)
}
