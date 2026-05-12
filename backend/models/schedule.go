package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScheduleItem 表示调度任务的结构
type ScheduleItem struct {
	ID           primitive.ObjectID   `bson:"_id" json:"id"`
	ScriptID     primitive.ObjectID   `bson:"script_id" json:"script_id"`
	UserID       primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Cron         string               `bson:"cron" json:"cron"`
	Enabled      bool                 `bson:"enabled" json:"enabled"`
	UseEnvGroups bool                 `bson:"use_env_groups" json:"use_env_groups"`
	EnvGroupIDs  []primitive.ObjectID `bson:"env_group_ids,omitempty" json:"env_group_ids,omitempty"`
	CreatedAt    time.Time            `bson:"created_at" json:"created_at"`
	NodeID       *primitive.ObjectID  `bson:"node_id,omitempty" json:"node_id,omitempty"`
}
