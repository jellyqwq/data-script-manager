package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jellyqwq/data-script-manager/backend/handlers"
	"github.com/jellyqwq/data-script-manager/backend/middleware"
)

func Setup(app *fiber.App) {
	api := app.Group("/api")

	// 不需要登录的接口
	api.Post("/login", handlers.Login)
	api.Post("/register", handlers.Register)
	api.Post("/reset-password", handlers.ResetPassword)
	api.Post("/send-code", handlers.SendCode)

	auth := api.Group("/auth", middleware.AuthRequired)

	auth.Get("/scripts", handlers.GetScripts)
	auth.Post("/scripts", handlers.CreateScript)
	auth.Put("/scripts/:id", handlers.UpdateScript)
	auth.Delete("/scripts/:id", handlers.DeleteScript)

	// 任务调度模块
	auth.Get("/schedules", handlers.GetSchedules)
	auth.Post("/schedules", handlers.AddSchedule)
	auth.Put("/schedules/:id", handlers.UpdateSchedule)
	auth.Delete("/schedules/:id", handlers.DeleteSchedule)

	auth.Get("/nodes", handlers.GetNodes)
	auth.Post("/nodes", handlers.AddNode)
	auth.Put("/nodes/:id", handlers.UpdateNode)
	auth.Delete("/nodes/:id", handlers.DeleteNode)

	auth.Get("/logs", handlers.GetLogs)
	auth.Delete("/logs/:id", handlers.DeleteLog)
	auth.Delete("/logs", handlers.ClearLogs)

	auth.Get("/env-vars", handlers.GetEnvVars)
	auth.Post("/env-vars", handlers.CreateEnvVar)
	auth.Put("/env-vars/:id", handlers.UpdateEnvVar)
	auth.Delete("/env-vars/:id", handlers.DeleteEnvVar)

}
