package utils

import "golang.org/x/crypto/bcrypt"

// 加密密码
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

// 验证密码（输入的明文密码 和 数据库中的 hash 对比）
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
