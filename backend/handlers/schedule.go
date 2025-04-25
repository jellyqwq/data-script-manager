package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/scheduler"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Schedule struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ScriptID  string             `bson:"script_id" json:"script_id"`
	Cron      string             `bson:"cron" json:"cron"`
	Enabled   bool               `bson:"enabled" json:"enabled"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	NextRun   time.Time          `bson:"next_run" json:"next_run"`
}

// 获取所有调度任务
func GetSchedules(c *fiber.Ctx) error {
	col := db.Mongo.Database("scriptdb").Collection("schedules")
	cursor, err := col.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "查询失败",
		})
	}
	var list []Schedule
	if err := cursor.All(context.TODO(), &list); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "数据解析失败",
		})
	}
	return c.JSON(list)
}

// 新增调度任务
func AddSchedule(c *fiber.Ctx) error {
	var input struct {
		ScriptID string `json:"script_id"`
		Cron     string `json:"cron"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "参数错误",
		})
	}

	// ✅ 解析 string 类型的 ScriptID
	scriptOID, err := primitive.ObjectIDFromHex(input.ScriptID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的脚本ID",
		})
	}

	sched := bson.M{
		"script_id":  scriptOID,
		"cron":       input.Cron,
		"enabled":    true,
		"created_at": time.Now(),
		"next_run":   time.Now().Add(1 * time.Hour),
	}

	col := db.Mongo.Database("scriptdb").Collection("schedules")
	result, err := col.InsertOne(context.TODO(), sched)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "任务添加失败",
		})
	}

	// ✅ 插入成功后立即刷新任务调度器
	insertedID := result.InsertedID.(primitive.ObjectID).Hex()
	scheduler.ReloadSchedule(insertedID)

	return c.JSON(fiber.Map{
		"message": "添加成功",
	})
}

// 删除调度任务
func DeleteSchedule(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的任务ID",
		})
	}

	col := db.Mongo.Database("scriptdb").Collection("schedules")
	_, err = col.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "任务删除失败",
		})
	}

	// ✅ 删除任务后，移除调度器中的定时任务
	scheduler.RemoveSchedule(id)

	return c.JSON(fiber.Map{
		"message": "删除成功",
	})
}

// 修改调度任务
func UpdateSchedule(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID格式错误")
	}

	var body struct {
		Cron    string `json:"cron"`
		Enabled *bool  `json:"enabled"` // 可选更新
	}

	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "参数解析失败")
	}

	update := bson.M{}
	if body.Cron != "" {
		update["cron"] = body.Cron
	}
	if body.Enabled != nil {
		update["enabled"] = *body.Enabled
	}
	if len(update) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "无有效字段更新")
	}

	col := db.Mongo.Database("scriptdb").Collection("schedules")
	_, err = col.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "数据库更新失败")
	}

	// ✅ 更新成功后刷新该调度器任务
	scheduler.ReloadSchedule(id)

	return c.JSON(fiber.Map{"message": "更新成功"})
}
