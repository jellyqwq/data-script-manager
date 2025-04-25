package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LogEntry struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ScriptID  primitive.ObjectID `bson:"script_id" json:"script_id"`
	UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id"` // 👈 新增字段
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	Level     string             `bson:"level" json:"level"` // INFO / ERROR / DEBUG
	Message   string             `bson:"message" json:"message"`
}
