package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type EnvVar struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`
	Key    string             `bson:"key" json:"key"`
	Value  string             `bson:"value" json:"value"`
}
