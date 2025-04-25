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
	log.Println("加载环境变量完成")

	db.ConnectRedis()  // 初始化Redis数据库
	log.Println("Redis初始化完成")

	db.ConnectMongo()
	log.Println("MongoDB初始化完成")

	scheduler.StartScheduler()
	log.Println("任务调度更新模块启动")

	db.StartNodeHeartbeat() // ✅ 开始定时资源上报
	log.Println("资源上报启动")

	app := fiber.New()
	app.Use(middleware.CORS())
	routes.Setup(app)

	log.Fatal(app.Listen(":8080"))
}
