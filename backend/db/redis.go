package db

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),     // e.g. "localhost:6379"
		Password: os.Getenv("REDIS_PASSWORD"), // 如果有密码
		DB:       0,                           // 默认使用第0个数据库
	})

	// 测试连接
	_, err := Redis.Ping(Ctx).Result()
	if err != nil {
		log.Panicf("Redis 连接失败: %v", err)
	}

	log.Println("Redis 连接成功")
}
