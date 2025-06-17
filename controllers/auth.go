package controllers

import (
	"context"
	"strings"
	"time"

	"sitor-backend/config"
	"sitor-backend/models"
	"sitor-backend/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCol = config.GetDB().Collection(config.UserCollection)

func Register(c *fiber.Ctx) error {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input"})
	}
	if input.Name == "" || input.Email == "" || input.Password == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "All fields required"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Cek email unik
	count, _ := userCol.CountDocuments(ctx, bson.M{"email": strings.ToLower(input.Email)})
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Email already registered"})
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	user := models.User{
		ID:           primitive.NewObjectID(),
		Name:         input.Name,
		Email:        strings.ToLower(input.Email),
		Password:     string(hash),
		JoinedGroups: []primitive.ObjectID{},
		CreatedAt:    time.Now(),
	}
	_, err := userCol.InsertOne(ctx, user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to register"})
	}
	token, _ := utils.GenerateJWT(user.ID.Hex(), user.Email, "netral")
	return c.JSON(fiber.Map{
		"success": true,
		"user": fiber.Map{
			"id":           user.ID.Hex(),
			"email":        user.Email,
			"name":         user.Name,
			"joinedGroups": user.JoinedGroups,
			"createdAt":    user.CreatedAt,
		},
		"token": token,
	})
}

func Login(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user models.User
	err := userCol.FindOne(ctx, bson.M{"email": strings.ToLower(input.Email)}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Invalid email or password"})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Server error"})
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)) != nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Invalid email or password"})
	}
	token, _ := utils.GenerateJWT(user.ID.Hex(), user.Email, "netral")
	return c.JSON(fiber.Map{
		"success": true,
		"user": fiber.Map{
			"id":           user.ID.Hex(),
			"email":        user.Email,
			"name":         user.Name,
			"joinedGroups": user.JoinedGroups,
			"createdAt":    user.CreatedAt,
		},
		"token": token,
	})
}

func Me(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user models.User
	objId, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid userId"})
	}
	err = userCol.FindOne(ctx, bson.M{"_id": objId}).Decode(&user)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "User not found"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"user": fiber.Map{
			"id":           user.ID.Hex(),
			"email":        user.Email,
			"name":         user.Name,
			"joinedGroups": user.JoinedGroups,
			"createdAt":    user.CreatedAt,
		},
	})
}
