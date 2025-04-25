package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/middleware"
	"github.com/jellyqwq/data-script-manager/backend/routes"
	"github.com/jellyqwq/data-script-manager/backend/config"
	"github.com/jellyqwq/data-script-manager/backend/scheduler"
)

func main() {
	log.Println("✅ 服务启动中...")
	config.LoadEnv()
	app := fiber.New()
	app.Use(middleware.CORS())

	db.ConnectMongo()
	scheduler.StartScheduler()
	db.StartNodeHeartbeat() // ✅ 开始定时资源上报

	routes.Setup(app)

	log.Fatal(app.Listen(":8080"))
}
