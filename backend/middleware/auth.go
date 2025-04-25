package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jellyqwq/data-script-manager/backend/utils"
)

func AuthRequired(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "未提供有效的身份验证令牌",
		})
	}

	tokenStr := strings.TrimPrefix(auth, "Bearer ")

	_, err := utils.ParseToken(tokenStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "无效或已过期的 token",
		})
	}

	return c.Next()
}
