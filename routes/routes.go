package routes

import (
	"sitor-backend/controllers"
	"sitor-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	// Auth
	api.Post("/register", controllers.Register)
	api.Post("/login", controllers.Login)
	api.Get("/me", middleware.JWTProtected(), controllers.Me)

	// Group
	api.Get("/groups", controllers.GetGroups)
	api.Post("/groups", middleware.JWTProtected(), controllers.CreateGroup)
	api.Post("/groups/join", middleware.JWTProtected(), controllers.JoinGroup)
	api.Get("/groups/:id/members", middleware.JWTProtected(), controllers.ListGroupMembers)

	// Detection
	api.Post("/detections", middleware.JWTProtected(), controllers.CreateDetection)
	api.Get("/detections/:groupId", middleware.JWTProtected(), controllers.GetDetectionsByGroup)
}
