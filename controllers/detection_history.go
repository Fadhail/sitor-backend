package controllers

import (
	"context"
	"sitor-backend/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GET /api/groups/:groupId/history
func GetDetectionHistory(c *fiber.Ctx) error {
	groupId := c.Params("groupId")
	if groupId == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Group ID required"})
	}
	objGroupId, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid groupId"})
	}
	db := config.GetDB()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := db.Collection("detection_history").Find(ctx, bson.M{"groupId": objGroupId})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to fetch history", "error": err.Error()})
	}
	var history []bson.M
	err = cursor.All(ctx, &history)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to parse history", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "history": history})
}
