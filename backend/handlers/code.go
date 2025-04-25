package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func SendCode(c *fiber.Ctx) error {
	type Request struct {
		Email string `json:"email"`
		Scene string `json:"scene"` // "register" 或 "reset"
	}

	var req Request
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "邮箱不能为空",
		})
	}

	if req.Scene != "register" && req.Scene != "reset" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "非法的验证码用途类型",
		})
	}

	collection := db.Mongo.Database("scriptdb").Collection("users")
	count, err := collection.CountDocuments(context.TODO(), bson.M{"username": req.Email})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "数据库查询失败",
		})
	}

	if req.Scene == "register" && count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "该邮箱已注册，请直接登录",
		})
	}

	if req.Scene == "reset" && count == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "该邮箱尚未注册",
		})
	}

	code := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	utils.SetCode(req.Email, code, 5*time.Minute)

	if err := utils.SendEmail(req.Email, code); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "验证码发送失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "验证码已发送",
	})
}
