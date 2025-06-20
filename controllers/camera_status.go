package controllers

import (
	"context"
	"fmt"
	"time"

	"sitor-backend/config"
	"sitor-backend/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cameraStatusCol = config.GetDB().Collection("camera_status")

// POST /api/groups/:groupId/camera-status
func UpdateCameraStatus(c *fiber.Ctx) error {
	groupId := c.Params("groupId")
	userId := c.Locals("userId")
	if groupId == "" || userId == nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Missing groupId or userId"})
	}
	var req struct {
		IsActive bool `json:"isActive"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid request body"})
	}
	gid, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid groupId"})
	}
	uid, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid userId"})
	}
	filter := bson.M{"groupId": gid, "userId": uid}
	update := bson.M{"$set": bson.M{"isActive": req.IsActive, "updatedAt": time.Now()}}
	opts := options.Update().SetUpsert(true)
	_, err = cameraStatusCol.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to update camera status"})
	}
	return c.JSON(fiber.Map{"success": true})
}

// GET /api/groups/:groupId/camera-status
func GetCameraStatus(c *fiber.Ctx) error {
	groupId := c.Params("groupId")
	if groupId == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Missing groupId"})
	}
	gid, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid groupId"})
	}

	// Cek status sesi grup
	db := config.GetDB()
	group := db.Collection("groups")
	var groupDoc struct {
		SessionActive bool `bson:"sessionActive"`
	}
	err = group.FindOne(context.Background(), bson.M{"_id": gid}).Decode(&groupDoc)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Group not found"})
	}
	if !groupDoc.SessionActive {
		return c.Status(410).JSON(fiber.Map{"success": false, "message": "Session has ended"})
	}

	// Ambil status kamera
	cursor, err := cameraStatusCol.Find(context.Background(), bson.M{"groupId": gid})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to fetch camera status"})
	}
	var statuses []models.CameraStatus
	if err := cursor.All(context.Background(), &statuses); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to decode camera status"})
	}
	// Tambahkan log untuk debug
	fmt.Println("[DEBUG] GetCameraStatus: sessionActive=", groupDoc.SessionActive, "jumlah status=", len(statuses))
	return c.JSON(fiber.Map{"success": true, "statuses": statuses})
}
