package handlers

import (
	"context"
	"time"
	"log"

	"github.com/gofiber/fiber/v2"
	
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/models"
)

type Node struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Address   string             `bson:"address" json:"address"`
	CPUUsage  float64            `bson:"cpu_usage" json:"cpu_usage"`
	MemUsage  float64            `bson:"mem_usage" json:"mem_usage"`
	DiskUsage float64            `bson:"disk_usage" json:"disk_usage"`
	Online    bool               `bson:"online" json:"online"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}


// 获取所有节点
func GetNodes(c *fiber.Ctx) error {
	col := db.Mongo.Database("scriptdb").Collection("nodes")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "查询失败"})
	}

	var nodes []models.Node
	if err := cursor.All(ctx, &nodes); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "解析失败"})
	}

	return c.JSON(nodes)
}

// 添加节点
func AddNode(c *fiber.Ctx) error {
	var node Node
	if err := c.BodyParser(&node); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "请求参数错误"})
	}
	node.UpdatedAt = time.Now()
	node.Online = true
	log.Println(node)

	col := db.Mongo.Database("scriptdb").Collection("nodes")
	_, err := col.InsertOne(context.TODO(), node)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "添加失败"})
	}
	return c.JSON(fiber.Map{"message": "添加成功"})
}

// 更新节点
func UpdateNode(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID格式错误"})
	}

	var updateData bson.M
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "请求体错误"})
	}
	updateData["updated_at"] = time.Now()

	col := db.Mongo.Database("scriptdb").Collection("nodes")
	_, err = col.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": updateData})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "更新失败"})
	}

	return c.JSON(fiber.Map{"message": "更新成功"})
}

// 删除节点
func DeleteNode(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID格式错误"})
	}

	col := db.Mongo.Database("scriptdb").Collection("nodes")
	_, err = col.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "删除失败"})
	}

	return c.JSON(fiber.Map{"message": "删除成功"})
}
