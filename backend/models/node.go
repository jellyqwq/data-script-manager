package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Node struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	Address    string             `bson:"address" json:"address"`
	CPUUsage   float64            `bson:"cpu_usage" json:"cpu_usage"`
	MemUsage   float64            `bson:"mem_usage" json:"mem_usage"`
	DiskUsage  float64            `bson:"disk_usage" json:"disk_usage"`
	Online     bool               `bson:"online" json:"online"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
