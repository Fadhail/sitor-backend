package main

import (
	"log"
	"sitor-backend/config"
	"sitor-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or failed to load .env")
	}

	app := fiber.New()

	// Tambahkan middleware recovery agar server tidak mati jika panic
	app.Use(recover.New())

	// Tambahkan middleware CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// Inisialisasi koneksi DB sekali saja
	_ = config.GetDB()

	routes.SetupRoutes(app)

	app.Listen(":8080")
}
