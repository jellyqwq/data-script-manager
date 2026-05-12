package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/models"
	"github.com/jellyqwq/data-script-manager/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type envGroupInput struct {
	ScriptID string           `json:"script_id"`
	Name     string           `json:"name"`
	Enabled  bool             `json:"enabled"`
	Vars     []models.EnvPair `json:"vars"`
}

func GetEnvGroups(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权"})
	}

	filter := bson.M{"user_id": userID}
	if scriptID := c.Query("script_id"); scriptID != "" {
		scriptOID, err := primitive.ObjectIDFromHex(scriptID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "无效的脚本ID"})
		}
		filter["script_id"] = scriptOID
	}

	col := db.Mongo.Database("scriptdb").Collection("env_groups")
	cursor, err := col.Find(
		context.TODO(),
		filter,
		options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}}),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "查询失败"})
	}
	defer cursor.Close(context.TODO())

	var groups []models.EnvGroup
	if err := cursor.All(context.TODO(), &groups); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "解析失败"})
	}

	return c.JSON(fiber.Map{"items": groups, "total": len(groups)})
}

func CreateEnvGroup(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权"})
	}

	var input envGroupInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "参数错误"})
	}

	scriptOID, err := parseAndVerifyScriptID(input.ScriptID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	now := time.Now()
	group := models.EnvGroup{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		ScriptID:  scriptOID,
		Name:      strings.TrimSpace(input.Name),
		Enabled:   input.Enabled,
		Vars:      cleanEnvPairs(input.Vars),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if group.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "变量组名称不能为空"})
	}

	col := db.Mongo.Database("scriptdb").Collection("env_groups")
	if _, err := col.InsertOne(context.TODO(), group); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "创建失败"})
	}

	return c.JSON(group)
}

func UpdateEnvGroup(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权"})
	}

	groupID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "无效的变量组ID"})
	}

	var input envGroupInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "参数错误"})
	}

	scriptOID, err := parseAndVerifyScriptID(input.ScriptID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "变量组名称不能为空"})
	}

	col := db.Mongo.Database("scriptdb").Collection("env_groups")
	result, err := col.UpdateOne(
		context.TODO(),
		bson.M{"_id": groupID, "user_id": userID},
		bson.M{"$set": bson.M{
			"script_id":  scriptOID,
			"name":       name,
			"enabled":    input.Enabled,
			"vars":       cleanEnvPairs(input.Vars),
			"updated_at": time.Now(),
		}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "更新失败"})
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "变量组不存在或无权限"})
	}

	return c.JSON(fiber.Map{"message": "更新成功"})
}

func DeleteEnvGroup(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "未授权"})
	}

	groupID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "无效的变量组ID"})
	}

	col := db.Mongo.Database("scriptdb").Collection("env_groups")
	result, err := col.DeleteOne(context.TODO(), bson.M{"_id": groupID, "user_id": userID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "删除失败"})
	}
	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "变量组不存在或无权限"})
	}

	return c.JSON(fiber.Map{"message": "删除成功"})
}

func parseAndVerifyScriptID(scriptID string, userID primitive.ObjectID) (primitive.ObjectID, error) {
	scriptOID, err := primitive.ObjectIDFromHex(scriptID)
	if err != nil {
		return primitive.NilObjectID, fiber.NewError(fiber.StatusBadRequest, "无效的脚本ID")
	}

	col := db.Mongo.Database("scriptdb").Collection("scripts")
	count, err := col.CountDocuments(context.TODO(), bson.M{"_id": scriptOID, "user_id": userID})
	if err != nil {
		return primitive.NilObjectID, err
	}
	if count == 0 {
		return primitive.NilObjectID, fiber.NewError(fiber.StatusForbidden, "脚本不存在或无权限")
	}

	return scriptOID, nil
}

func cleanEnvPairs(pairs []models.EnvPair) []models.EnvPair {
	cleaned := make([]models.EnvPair, 0, len(pairs))
	for _, pair := range pairs {
		key := strings.TrimSpace(pair.Key)
		if key == "" {
			continue
		}
		cleaned = append(cleaned, models.EnvPair{
			Key:   key,
			Value: pair.Value,
		})
	}
	return cleaned
}
