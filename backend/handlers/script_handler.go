package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/models"
	"github.com/jellyqwq/data-script-manager/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetScripts(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}
	collection := db.Mongo.Database("scriptdb").Collection("scripts")

	cur, err := collection.Find(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "数据库查询失败"})
	}

	var results []models.Script
	if err = cur.All(context.TODO(), &results); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "数据解析失败"})
	}

	return c.JSON(results)
}

func CreateScript(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}

	var input models.Script
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	input.UserID = userID
	input.CreatedAt = now
	input.LastModified = now

	collection := db.Mongo.Database("scriptdb").Collection("scripts")
	res, err := collection.InsertOne(context.TODO(), input)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "创建失败"})
	}

	input.ID = res.InsertedID.(primitive.ObjectID)
	return c.JSON(input)
}

func UpdateScript(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}
	scriptID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID 无效"})
	}

	var input models.Script
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
	}

	collection := db.Mongo.Database("scriptdb").Collection("scripts")
	filter := bson.M{"_id": scriptID, "user_id": userID}
	update := bson.M{"$set": bson.M{
		"script_name":   input.ScriptName,
		"description":   input.Description,
		"content":       input.Content,
		"last_modified": primitive.NewDateTimeFromTime(time.Now()),
	}}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "更新失败"})
	}

	return c.JSON(fiber.Map{"message": "更新成功"})
}

func DeleteScript(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}
	scriptID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID 无效"})
	}

	collection := db.Mongo.Database("scriptdb").Collection("scripts")
	filter := bson.M{"_id": scriptID, "user_id": userID}
	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "删除失败"})
	}

	return c.JSON(fiber.Map{"message": "删除成功"})
}
