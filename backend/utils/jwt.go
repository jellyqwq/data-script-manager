package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jellyqwq/data-script-manager/backend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtKey = []byte("secret")

// 生成 Token
func GenerateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ✅ 解析 Token
func ParseToken(tokenStr string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
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

// ✅ 新增：从 fiber 请求中提取 user_id，并转为 ObjectID
func GetUserIDFromToken(c *fiber.Ctx) primitive.ObjectID {
	auth := c.Get("Authorization")
	tokenStr := strings.TrimPrefix(auth, "Bearer ")

	claims, err := ParseToken(tokenStr)
	if err != nil {
		return primitive.NilObjectID
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return primitive.NilObjectID
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return primitive.NilObjectID
	}

	return userID
}
