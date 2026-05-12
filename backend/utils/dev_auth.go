package utils

import (
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const defaultDevUserID = "000000000000000000000001"

func IsDevAuthBypassEnabled() bool {
	switch strings.ToLower(os.Getenv("DEV_AUTH_BYPASS")) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func DevUserID() primitive.ObjectID {
	id := os.Getenv("DEV_AUTH_BYPASS_USER_ID")
	if id == "" {
		id = defaultDevUserID
	}

	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID
	}

	return userID
}
