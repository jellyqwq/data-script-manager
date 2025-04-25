package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/models"
	"github.com/jellyqwq/data-script-manager/backend/utils"
)

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "参数解析失败",
		})
	}

	collection := db.Mongo.Database("scriptdb").Collection("users")
	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{"username": input.Username}).Decode(&user)
	if err != nil || !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "用户名或密码错误",
		})
	}

	token, err := utils.GenerateToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Token 生成失败",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

func Register(c *fiber.Ctx) error {
	type Input struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Password string `json:"password"`
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "参数解析失败",
		})
	}
	if input.Email == "" || input.Code == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "请完整填写注册信息",
		})
	}
	if !utils.VerifyCode(input.Email, input.Code) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "验证码无效或已过期",
		})
	}

	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "密码加密失败",
		})
	}

	collection := db.Mongo.Database("scriptdb").Collection("users")
	_, err = collection.InsertOne(context.TODO(), models.User{
		Username:     input.Email,
		PasswordHash: hash,
		Role:         "user",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "注册失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "注册成功",
	})
}

func ResetPassword(c *fiber.Ctx) error {
	type Input struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Password string `json:"password"`
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "参数解析失败",
		})
	}
	if input.Email == "" || input.Code == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "请完整填写信息",
		})
	}
	if !utils.VerifyCode(input.Email, input.Code) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "验证码无效或已过期",
		})
	}

	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "密码加密失败",
		})
	}

	collection := db.Mongo.Database("scriptdb").Collection("users")
	filter := bson.M{"username": input.Email}
	update := bson.M{"$set": bson.M{"password_hash": hash}}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "密码重置失败",
		})
	}

	return c.JSON(fiber.Map{
		"message": "密码重置成功",
	})
}
