package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskRun struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	ScheduleID   primitive.ObjectID  `bson:"schedule_id" json:"schedule_id"`
	ScriptID     primitive.ObjectID  `bson:"script_id" json:"script_id"`
	UserID       primitive.ObjectID  `bson:"user_id" json:"user_id"`
	EnvGroupID   *primitive.ObjectID `bson:"env_group_id,omitempty" json:"env_group_id,omitempty"`
	EnvGroupName string              `bson:"env_group_name,omitempty" json:"env_group_name,omitempty"`
	StateDir     string              `bson:"state_dir,omitempty" json:"state_dir,omitempty"`
	Status       string              `bson:"status" json:"status"`
	ExitCode     *int                `bson:"exit_code,omitempty" json:"exit_code,omitempty"`
	ErrorMessage string              `bson:"error_message,omitempty" json:"error_message,omitempty"`
	StartedAt    time.Time           `bson:"started_at" json:"started_at"`
	FinishedAt   *time.Time          `bson:"finished_at,omitempty" json:"finished_at,omitempty"`
	DurationMs   int64               `bson:"duration_ms,omitempty" json:"duration_ms,omitempty"`
}
