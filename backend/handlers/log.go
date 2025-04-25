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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 日志分页 + 过滤（带用户归属校验）
func GetLogs(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromToken(c) // 👈 假设你有utils方法拿token里的用户ID
	if userID == primitive.NilObjectID {
		return c.Status(401).JSON(fiber.Map{"error": "无效的用户身份"})
	}

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
		scriptObjID, err := primitive.ObjectIDFromHex(scriptID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "无效脚本ID"})
		}
		// 校验脚本归属
		col := db.Mongo.Database("scriptdb").Collection("scripts")
		var script models.Script
		err = col.FindOne(context.TODO(), bson.M{"_id": scriptObjID, "user_id": userID}).Decode(&script)
		if err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "无权限查看该脚本日志"})
		}

		filter["script_id"] = scriptObjID
	} else {
		// 如果没筛选具体脚本，可以根据user_id过滤（可选）
		filter["user_id"] = userID
	}

	if level != "" {
		filter["level"] = level
	}

	logCol := db.Mongo.Database("scriptdb").Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	total, err := logCol.CountDocuments(ctx, filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "日志统计失败"})
	}

	opts := options.Find().
		SetLimit(int64(pageSize)).
		SetSkip(int64((page - 1) * pageSize)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := logCol.Find(ctx, filter, opts)
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

// 删除单条日志（增加归属校验）
func DeleteLog(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromToken(c)
	if userID == primitive.NilObjectID {
		return c.Status(401).JSON(fiber.Map{"error": "无效的用户身份"})
	}

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "无效日志ID"})
	}

	logCol := db.Mongo.Database("scriptdb").Collection("logs")

	// 先找到对应日志
	var log models.LogEntry
	err = logCol.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&log)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "日志不存在"})
	}

	// 校验对应脚本是否属于自己
	scriptCol := db.Mongo.Database("scriptdb").Collection("scripts")
	var script models.Script
	err = scriptCol.FindOne(context.TODO(), bson.M{"_id": log.ScriptID, "user_id": userID}).Decode(&script)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "无权限删除该日志"})
	}

	// 真的可以删了
	_, err = logCol.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "删除失败"})
	}

	return c.JSON(fiber.Map{"message": "日志已删除"})
}

// 清空所有日志（只清自己的）
func ClearLogs(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromToken(c)
	if userID == primitive.NilObjectID {
		return c.Status(401).JSON(fiber.Map{"error": "无效的用户身份"})
	}

	logCol := db.Mongo.Database("scriptdb").Collection("logs")

	// 只清属于自己的脚本产生的日志
	_, err := logCol.DeleteMany(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "清空失败"})
	}

	return c.JSON(fiber.Map{"message": "您的日志已全部清空"})
}
