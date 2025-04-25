package utils

import (
	"sync"
	"time"
)

type CodeEntry struct {
	Code      string
	ExpiresAt time.Time
}

var codeStore = make(map[string]CodeEntry)
var mu sync.Mutex

// 设置验证码（覆盖旧的）
func SetCode(email string, code string, duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	codeStore[email] = CodeEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(duration),
	}
}

// 校验验证码是否正确
func VerifyCode(email string, code string) bool {
	mu.Lock()
	defer mu.Unlock()
	entry, exists := codeStore[email]
	if !exists {
		return false
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(codeStore, email)
		return false
	}
	if entry.Code != code {
		return false
	}
	delete(codeStore, email) // 验证成功后删除（一次性）
	return true
}
