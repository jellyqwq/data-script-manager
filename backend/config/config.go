package config

import (
	"log"
	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("未加载 .env，使用系统环境变量")
	}
}
