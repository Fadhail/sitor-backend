package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sitor-backend/models"
)

var chatHistoryCollection *mongo.Collection

func InitChatHistoryCollection(db *mongo.Database) {
	chatHistoryCollection = db.Collection("chat_histories")
}

// GET /chat-history
func GetChatHistory(c *fiber.Ctx) error {
	userID := c.Locals("userId")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	uid, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid userId"})
	}
	var history models.ChatHistory
	err = chatHistoryCollection.FindOne(context.Background(), bson.M{"user_id": uid}).Decode(&history)
	if err == mongo.ErrNoDocuments {
		return c.JSON(fiber.Map{"messages": []models.ChatMessage{}})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"messages": history.Messages})
}

// POST /chat-history
func AddChatMessage(c *fiber.Ctx) error {
	userID := c.Locals("userId")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	uid, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid userId"})
	}
	var msg models.ChatMessage
	if err := c.BodyParser(&msg); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	msg.CreatedAt = time.Now()
	filter := bson.M{"user_id": uid}
	update := bson.M{"$push": bson.M{"messages": msg}}
	upsert := true
	_, err = chatHistoryCollection.UpdateOne(context.Background(), filter, update, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Message added"})
}
