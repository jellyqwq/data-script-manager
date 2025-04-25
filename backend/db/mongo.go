package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Mongo *mongo.Client

// 数据库连接
func ConnectMongo() {
  
  uri := fmt.Sprintf(
    "mongodb://%s:%s@%s:%s",
    os.Getenv("MONGODB_USER"),
    os.Getenv("MONGODB_PASSWORD"),
    os.Getenv("MONGODB_HOST"),
    os.Getenv("MONGODB_PORT"),
  )
  
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
  if err != nil {
    log.Fatal("MongoDB 连接失败:", err)
  }

  Mongo = client
  log.Println("MongoDB 连接成功")
}
