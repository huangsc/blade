package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthenticator JWT认证器
type JWTAuthenticator struct {
	// secretKey 密钥
	secretKey []byte
	// expiration 过期时间
	expiration time.Duration
}

// JWTClaims JWT声明
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// NewJWTAuthenticator 创建JWT认证器
func NewJWTAuthenticator(secretKey string, expiration time.Duration) *JWTAuthenticator {
	return &JWTAuthenticator{
		secretKey:  []byte(secretKey),
		expiration: expiration,
	}
}

// GenerateToken 生成JWT令牌
func (a *JWTAuthenticator) GenerateToken(claims Claims) (string, error) {
	now := time.Now()
	jwtClaims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(a.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString(a.secretKey)
}

// ValidateToken 验证JWT令牌
func (a *JWTAuthenticator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	jwtClaims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	// 检查令牌是否过期
	if jwtClaims.ExpiresAt.Unix() < time.Now().Unix() {
		return nil, ErrExpiredToken
	}

	return &Claims{
		UserID:    jwtClaims.UserID,
		Username:  jwtClaims.Username,
		Role:      jwtClaims.Role,
		ExpiresAt: jwtClaims.ExpiresAt.Unix(),
		IssuedAt:  jwtClaims.IssuedAt.Unix(),
	}, nil
}
