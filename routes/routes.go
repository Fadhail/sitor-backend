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
	api.Get("/me/summary", middleware.JWTProtected(), controllers.MeSummary)
	api.Patch("/me", middleware.JWTProtected(), controllers.UpdateProfile)
	api.Patch("/me/password", middleware.JWTProtected(), controllers.UpdatePassword)

	// Group
	api.Get("/groups", controllers.GetGroups)
	api.Post("/groups", middleware.JWTProtected(), controllers.CreateGroup)
	api.Post("/groups/join", middleware.JWTProtected(), controllers.JoinGroup)
	api.Get("/groups/:id/members", middleware.JWTProtected(), controllers.ListGroupMembers)
	api.Delete("/groups/:id", middleware.JWTProtected(), controllers.DeleteGroup)
	api.Post("/groups/:id/leave", middleware.JWTProtected(), controllers.LeaveGroup)
	// End session (disconnect all users in group, but keep group)
	api.Post("/groups/:groupId/end-session", middleware.JWTProtected(), controllers.EndSession)
	// Start session (activate sessionActive on group)
	api.Post("/groups/:groupId/start-session", middleware.JWTProtected(), controllers.StartSession)

	// Detection
	api.Post("/detections", middleware.JWTProtected(), controllers.CreateDetection)
	api.Get("/detections/:groupId", middleware.JWTProtected(), controllers.GetDetectionsByGroup)
	// Detection history (riwayat sesi)
	api.Get("/groups/:groupId/history", middleware.JWTProtected(), controllers.GetDetectionHistory)

	// Camera status
	api.Post("/groups/:groupId/camera-status", middleware.JWTProtected(), controllers.UpdateCameraStatus)
	api.Get("/groups/:groupId/camera-status", middleware.JWTProtected(), controllers.GetCameraStatus)

	// Chat history
	api.Get("/chat-history", middleware.JWTProtected(), controllers.GetChatHistory)
	api.Post("/chat-history", middleware.JWTProtected(), controllers.AddChatMessage)
}
