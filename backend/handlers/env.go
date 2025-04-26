package handlers

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/models"
	"github.com/jellyqwq/data-script-manager/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GET /auth/env-vars
func GetEnvVars(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}

	pageStr := c.Query("page", "1")
	pageSizeStr := c.Query("pageSize", "10")
	sortBy := c.Query("sortBy", "key")       // 默认按 key 排序
	sortOrder := c.Query("sortOrder", "asc") // 默认升序

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	skip := (page - 1) * pageSize

	collection := db.Mongo.Database("scriptdb").Collection("env_vars")

	// 查询总数
	count, err := collection.CountDocuments(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "获取总数失败"})
	}

	// 构建排序选项
	sortOption := 1 // 1 for ascending, -1 for descending
	if sortOrder == "desc" {
		sortOption = -1
	}
	sort := bson.D{{Key: sortBy, Value: sortOption}}

	// 查询分页和排序后的数据
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize)).SetSort(sort)
	cur, err := collection.Find(context.TODO(), bson.M{"user_id": userID}, findOptions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "数据库查询失败"})
	}
	defer cur.Close(context.TODO())

	var results []models.EnvVar
	if err = cur.All(context.TODO(), &results); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "数据解析失败"})
	}

	return c.JSON(fiber.Map{
		"items": results,
		"total": count,
	})
}

// POST /auth/env-vars
func CreateEnvVar(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromToken(c)
	if userID == primitive.NilObjectID {
		return c.Status(401).JSON(fiber.Map{"error": "无效的用户身份"})
	}

	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	c.BodyParser(&body)
	col := db.Mongo.Database("scriptdb").Collection("env_vars")
	_, err := col.InsertOne(context.TODO(), bson.M{
		"user_id": userID,
		"key":     body.Key,
		"value":   body.Value,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "新增失败"})
	}
	return c.JSON(fiber.Map{"message": "已添加"})
}

// PUT /auth/env-vars/:id
func UpdateEnvVar(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID 无效"})
	}

	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "请求体解析失败"})
	}

	col := db.Mongo.Database("scriptdb").Collection("env_vars")
	result, err := col.UpdateOne(
		context.TODO(),
		bson.M{"_id": id, "user_id": userID}, // 同时匹配 _id 和 user_id
		bson.M{"$set": bson.M{
			"key":   body.Key,
			"value": body.Value,
		}},
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "更新失败"})
	}

	if result.ModifiedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "未找到要更新的记录或无权限修改"})
	}

	return c.JSON(fiber.Map{"message": "更新成功"})
}

// DELETE /auth/env-vars/:id
func DeleteEnvVar(c *fiber.Ctx) error {
	id, _ := primitive.ObjectIDFromHex(c.Params("id"))
	col := db.Mongo.Database("scriptdb").Collection("env_vars")
	_, err := col.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "删除失败"})
	}
	return c.JSON(fiber.Map{"message": "已删除"})
}
