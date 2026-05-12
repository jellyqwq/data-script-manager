package handlers

import (
	"context"
	"strconv"
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
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetProjection(bson.M{"content": 0}).
		SetSort(bson.D{{Key: "last_modified", Value: -1}})
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

	scriptID := primitive.NewObjectID()
	now := primitive.NewDateTimeFromTime(time.Now())
	input := models.Script{
		ID:           scriptID,
		UserID:       userID,
		CreatedAt:    now,
		LastModified: now,
	}

	if isMultipart(c) {
		fileHeader, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "请上传脚本文件"})
		}

		meta, err := utils.SaveUploadedScriptFile(userID, scriptID, fileHeader)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		input.ScriptName = c.FormValue("script_name")
		if input.ScriptName == "" {
			input.ScriptName = strings.TrimSuffix(meta.OriginalFilename, "."+fileExt(meta.OriginalFilename))
		}
		input.Description = c.FormValue("description")
		applyScriptFileMeta(&input, meta)
	} else {
		if err := c.BodyParser(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
		}
		input.ID = scriptID
		input.UserID = userID
		input.CreatedAt = now
		input.LastModified = now

		if input.Content == "" {
			return c.Status(400).JSON(fiber.Map{"error": "脚本内容不能为空"})
		}

		filename := input.ScriptName
		if filename == "" {
			filename = "script.py"
		}
		if !strings.Contains(filename, ".") {
			filename += ".py"
		}

		meta, err := utils.SaveScriptContent(userID, scriptID, filename, []byte(input.Content))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		input.Content = ""
		applyScriptFileMeta(&input, meta)
	}

	if input.ScriptName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "脚本名称不能为空"})
	}

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

	collection := db.Mongo.Database("scriptdb").Collection("scripts")
	filter := bson.M{"_id": scriptID, "user_id": userID}

	var oldScript models.Script
	if err := collection.FindOne(context.TODO(), filter).Decode(&oldScript); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "脚本不存在或无权限"})
	}

	setFields := bson.M{
		"last_modified": primitive.NewDateTimeFromTime(time.Now()),
	}

	if isMultipart(c) {
		if name := c.FormValue("script_name"); name != "" {
			setFields["script_name"] = name
		}
		setFields["description"] = c.FormValue("description")

		if fileHeader, err := c.FormFile("file"); err == nil && fileHeader != nil {
			meta, err := utils.SaveUploadedScriptFile(userID, scriptID, fileHeader)
			if err != nil {
				return c.Status(400).JSON(fiber.Map{"error": err.Error()})
			}
			utils.RemoveScriptFile(oldScript.FilePath)
			applyScriptFileMetaToSet(setFields, meta)
			setFields["content"] = ""
		}
	} else {
		var input models.Script
		if err := c.BodyParser(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "参数错误"})
		}
		setFields["script_name"] = input.ScriptName
		setFields["description"] = input.Description
		if input.Content != "" {
			filename := input.ScriptName
			if filename == "" {
				filename = oldScript.ScriptName
			}
			if !strings.Contains(filename, ".") {
				filename += ".py"
			}
			meta, err := utils.SaveScriptContent(userID, scriptID, filename, []byte(input.Content))
			if err != nil {
				return c.Status(400).JSON(fiber.Map{"error": err.Error()})
			}
			utils.RemoveScriptFile(oldScript.FilePath)
			applyScriptFileMetaToSet(setFields, meta)
			setFields["content"] = ""
		}
	}

	update := bson.M{"$set": setFields}
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
	var script models.Script
	if err := collection.FindOne(context.TODO(), filter).Decode(&script); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "脚本不存在或无权限"})
	}

	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "删除失败"})
	}
	utils.RemoveScriptFile(script.FilePath)

	return c.JSON(fiber.Map{"message": "删除成功"})
}

func isMultipart(c *fiber.Ctx) bool {
	return strings.Contains(strings.ToLower(c.Get("Content-Type")), "multipart/form-data")
}

func applyScriptFileMeta(script *models.Script, meta utils.ScriptFileMeta) {
	script.OriginalFilename = meta.OriginalFilename
	script.FilePath = meta.FilePath
	script.SHA1 = meta.SHA1
	script.Size = meta.Size
	script.Language = meta.Language
}

func applyScriptFileMetaToSet(setFields bson.M, meta utils.ScriptFileMeta) {
	setFields["original_filename"] = meta.OriginalFilename
	setFields["file_path"] = meta.FilePath
	setFields["sha1"] = meta.SHA1
	setFields["size"] = meta.Size
	setFields["language"] = meta.Language
}

func fileExt(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) <= 1 {
		return ""
	}
	return parts[len(parts)-1]
}
