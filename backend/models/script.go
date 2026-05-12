package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Script struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	ScriptName       string             `bson:"script_name" json:"script_name"`
	Description      string             `bson:"description" json:"description"`
	Content          string             `bson:"content,omitempty" json:"content,omitempty"`
	OriginalFilename string             `bson:"original_filename,omitempty" json:"original_filename,omitempty"`
	FilePath         string             `bson:"file_path,omitempty" json:"file_path,omitempty"`
	SHA1             string             `bson:"sha1,omitempty" json:"sha1,omitempty"`
	Size             int64              `bson:"size,omitempty" json:"size,omitempty"`
	Language         string             `bson:"language,omitempty" json:"language,omitempty"`
	CreatedAt        primitive.DateTime `bson:"created_at" json:"created_at"`
	LastModified     primitive.DateTime `bson:"last_modified" json:"last_modified"`
}
