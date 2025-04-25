package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/jellyqwq/data-script-manager/backend/db"
)

// 设置验证码
func SetCode(email string, code string, duration time.Duration) error {
	key := fmt.Sprintf("code:%s", email)
	return db.Redis.Set(context.Background(), key, code, duration).Err()
}

// 校验验证码
func VerifyCode(email string, code string) bool {
	key := fmt.Sprintf("code:%s", email)
	val, err := db.Redis.Get(context.Background(), key).Result()
	if err != nil || val != code {
		return false
	}
	// 验证成功后删除（一次性）
	db.Redis.Del(context.Background(), key)
	return true
}
