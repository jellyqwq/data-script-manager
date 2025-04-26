package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/models"
	"github.com/jellyqwq/data-script-manager/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetScripts(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}

	pageStr := c.Query("page", "1")
	pageSizeStr := c.Query("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10 // 设置一个合理的默认值和上限
	}

	skip := (page - 1) * pageSize

	collection := db.Mongo.Database("scriptdb").Collection("scripts")

	// 查询总数
	count, err := collection.CountDocuments(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "获取总数失败"})
	}

	// 查询分页数据
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize))
	cur, err := collection.Find(context.TODO(), bson.M{"user_id": userID}, findOptions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "数据库查询失败"})
	}
	defer cur.Close(context.TODO())

	var results []models.Script
	if err = cur.All(context.TODO(), &results); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "数据解析失败"})
	}

	return c.JSON(fiber.Map{
		"items": results,
		"total": count,
	})
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
