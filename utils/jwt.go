package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JwtClaims JWT载荷
type JwtClaims struct {
	UserID   uint64 `json:"user_id"`
	UserType int    `json:"user_type"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID uint64, userType int, secret string, expireHours int) (string, error) {
	// 构造载荷
	claims := JwtClaims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expireHours))), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                             // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                                             // 生效时间
		},
	}

	// 生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析JWT token
func ParseToken(tokenStr string, secret string) (*JwtClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenStr, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	// 验证token并获取载荷
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
