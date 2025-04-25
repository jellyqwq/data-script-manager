package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Script struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	ScriptName   string             `bson:"script_name" json:"script_name"`
	Description  string             `bson:"description" json:"description"`
	Content      string             `bson:"content" json:"content"`
	CreatedAt    primitive.DateTime `bson:"created_at" json:"created_at"`
	LastModified primitive.DateTime `bson:"last_modified" json:"last_modified"`
}
