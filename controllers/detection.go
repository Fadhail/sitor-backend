package controllers

import (
	"log"
	"sitor-backend/config"
	"sitor-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var detectionCol = config.GetDB().Collection("detections")

// POST /api/detections
func CreateDetection(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in CreateDetection: %v", r)
			c.Status(500).JSON(fiber.Map{"success": false, "message": "Internal server error (panic)"})
		}
	}()
	var body struct {
		GroupId     string         `json:"groupId"`
		Emotion     string         `json:"emotion"`
		Probability float64        `json:"probability"`
		Emotions    models.Emotion `json:"emotions"`
	}
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized: userId missing"})
	}
	userIdStr, ok := userId.(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized: userId not string"})
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid request"})
	}
	objGroupId, err := primitive.ObjectIDFromHex(body.GroupId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid groupId"})
	}
	objUserId, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid userId"})
	}
	// Ambil tanggal hari ini dalam format YYYY-MM-DD
	dateStr := time.Now().Format("2006-01-02")
	col := config.GetDB().Collection("detections")
	// Ambil nama user dari database
	userCol := config.GetDB().Collection("users")
	var user models.User
	err = userCol.FindOne(c.Context(), bson.M{"_id": objUserId}).Decode(&user)
	if err != nil {
		log.Printf("[CreateDetection] Failed to get user name: %v", err)
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to get user name"})
	}
	// Upsert: hanya satu data deteksi per user per grup
	filter := bson.M{
		"groupId": objGroupId,
		"userId":  objUserId,
	}
	if body.Emotions != (models.Emotion{}) {
		update := bson.M{
			"$set": bson.M{
				"groupId":   objGroupId,
				"userId":    objUserId,
				"userName":  user.Name,
				"timestamp": time.Now(),
				"date":      dateStr,
				"emotions":  body.Emotions,
			},
		}
		opts := options.Update().SetUpsert(true)
		_, err = col.UpdateOne(c.Context(), filter, update, opts)
		if err != nil {
			log.Printf("[CreateDetection] Failed to upsert detection: %v", err)
			return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to save detection"})
		}
		return c.JSON(fiber.Map{"success": true})
	}
	// fallback: jika tidak ada field emotions, gunakan logika lama (1/0)
	emotionUpdate := bson.M{
		"neutral":   0,
		"happy":     0,
		"sad":       0,
		"angry":     0,
		"surprised": 0,
		"disgusted": 0,
	}
	switch body.Emotion {
	case "neutral":
		emotionUpdate["neutral"] = 1
	case "happy":
		emotionUpdate["happy"] = 1
	case "sad":
		emotionUpdate["sad"] = 1
	case "angry":
		emotionUpdate["angry"] = 1
	case "surprised":
		emotionUpdate["surprised"] = 1
	case "disgusted":
		emotionUpdate["disgusted"] = 1
	}
	update := bson.M{
		"$set": bson.M{
			"groupId":   objGroupId,
			"userId":    objUserId,
			"userName":  user.Name,
			"timestamp": time.Now(),
			"date":      dateStr,
			"emotions":  emotionUpdate,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = col.UpdateOne(c.Context(), filter, update, opts)
	if err != nil {
		log.Printf("[CreateDetection] Failed to upsert detection: %v", err)
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to save detection"})
	}
	return c.JSON(fiber.Map{"success": true})
}

// GET /api/detections/:groupId
func GetDetectionsByGroup(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in GetDetectionsByGroup: %v", r)
			c.Status(500).JSON(fiber.Map{"success": false, "message": "Internal server error (panic)"})
		}
	}()
	groupId := c.Params("groupId")
	objGroupId, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid groupId"})
	}
	col := config.GetDB().Collection("detections")
	cursor, err := col.Find(c.Context(), bson.M{"groupId": objGroupId})
	if err != nil {
		log.Printf("[GetDetectionsByGroup] Failed to fetch detections: %v", err)
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to fetch detections"})
	}
	var detections []models.Detection
	if err := cursor.All(c.Context(), &detections); err != nil {
		log.Printf("[GetDetectionsByGroup] Failed to decode detections: %v", err)
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to decode detections"})
	}
	return c.JSON(fiber.Map{"success": true, "detections": detections})
}
