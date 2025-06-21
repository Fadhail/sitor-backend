package main

import (
	"log"
	"sitor-backend/config"
	"sitor-backend/controllers"
	"sitor-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file dari path pasti
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found or failed to load .env")
	}

	app := fiber.New()

	// Tambahkan middleware CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000, https://xeroon.xyz",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// Inisialisasi koneksi DB sekali saja
	db := config.GetDB()
	controllers.InitChatHistoryCollection(db)

	routes.SetupRoutes(app)

	app.Listen(":8080")
}
