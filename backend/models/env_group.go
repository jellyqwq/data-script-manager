package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EnvPair struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
}

type EnvGroup struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	ScriptID  primitive.ObjectID `bson:"script_id" json:"script_id"`
	Name      string             `bson:"name" json:"name"`
	Enabled   bool               `bson:"enabled" json:"enabled"`
	Vars      []EnvPair          `bson:"vars" json:"vars"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
