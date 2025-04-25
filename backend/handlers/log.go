package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 日志分页 + 过滤
func GetLogs(c *fiber.Ctx) error {
	scriptID := c.Query("script_id", "")
	level := c.Query("level", "")
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 50)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	filter := bson.M{}
	if scriptID != "" {
		if id, err := primitive.ObjectIDFromHex(scriptID); err == nil {
			filter["script_id"] = id
		}
	}
	if level != "" {
		filter["level"] = level
	}

	col := db.Mongo.Database("scriptdb").Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取总数
	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "日志统计失败"})
	}

	opts := options.Find().
		SetLimit(int64(pageSize)).
		SetSkip(int64((page - 1) * pageSize)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "日志获取失败"})
	}
	defer cursor.Close(ctx)

	var logs []models.LogEntry
	if err := cursor.All(ctx, &logs); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "数据解析失败"})
	}

	return c.JSON(fiber.Map{
		"data":  logs,
		"total": total,
	})
}

// 删除单条日志
func DeleteLog(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "无效日志 ID"})
	}
	col := db.Mongo.Database("scriptdb").Collection("logs")
	_, err = col.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "删除失败"})
	}
	return c.JSON(fiber.Map{"message": "日志已删除"})
}

// 清空所有日志
func ClearLogs(c *fiber.Ctx) error {
	col := db.Mongo.Database("scriptdb").Collection("logs")
	_, err := col.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "清空失败"})
	}
	return c.JSON(fiber.Map{"message": "所有日志已清空"})
}
