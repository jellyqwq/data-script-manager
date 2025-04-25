package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jellyqwq/data-script-manager/backend/models"
)

var jwtKey = []byte("secret")

// 生成 Token（你原有的逻辑）
func GenerateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ✅ 解析 Token（新增部分）
func ParseToken(tokenStr string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// 验证签名算法是否一致
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return map[string]interface{}{
			"user_id": claims["user_id"],
			"role":    claims["role"],
			"exp":     claims["exp"],
		}, nil
	}

	return nil, errors.New("invalid token")
}
