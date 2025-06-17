package controllers

import (
	"context"
	"sitor-backend/config"
	"sitor-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var chatCol = config.GetDB().Collection("chats")

func CreateChat(c *fiber.Ctx) error {
	groupId := c.Params("id")
	userId := c.Locals("userId")
	role := c.Locals("role")
	var body struct {
		Message  string `json:"message"`
		IsFromAI bool   `json:"isFromAI"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid request"})
	}
	objGroupId, _ := primitive.ObjectIDFromHex(groupId)
	var userObjId *primitive.ObjectID
	var userName string
	if !body.IsFromAI {
		id, err := primitive.ObjectIDFromHex(userId.(string))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid userId"})
		}
		userObjId = &id
		userName = c.Locals("name").(string)
	} else {
		userObjId = nil
		userName = "AI"
	}
	chat := models.ChatMessage{
		ID:        primitive.NewObjectID(),
		GroupID:   objGroupId,
		UserID:    userObjId,
		UserName:  userName,
		Role:      role.(string),
		Message:   body.Message,
		Timestamp: time.Now(),
		IsFromAI:  body.IsFromAI,
	}
	_, err := chatCol.InsertOne(context.TODO(), chat)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to save chat"})
	}
	return c.JSON(fiber.Map{"success": true, "chat": chat})
}

func ListChats(c *fiber.Ctx) error {
	groupId := c.Params("id")
	objGroupId, _ := primitive.ObjectIDFromHex(groupId)
	cur, err := chatCol.Find(context.TODO(), bson.M{"groupId": objGroupId})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to get chats"})
	}
	var chats []models.ChatMessage
	if err := cur.All(context.TODO(), &chats); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to parse chats"})
	}
	return c.JSON(fiber.Map{"success": true, "chats": chats})
}
