package utils

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ExtractUserID(c *fiber.Ctx) (primitive.ObjectID, error) {
	if IsDevAuthBypassEnabled() {
		userID := DevUserID()
		if userID == primitive.NilObjectID {
			return primitive.NilObjectID, errors.New("开发用户 ID 无效")
		}
		return userID, nil
	}

	auth := c.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return primitive.NilObjectID, errors.New("无 token")
	}
	token := auth[len("Bearer "):]
	claims, err := ParseToken(token)
	if err != nil {
		return primitive.NilObjectID, err
	}

	str, ok := claims["user_id"].(string)
	if !ok {
		return primitive.NilObjectID, errors.New("无效 user_id")
	}
	return primitive.ObjectIDFromHex(str)
}
