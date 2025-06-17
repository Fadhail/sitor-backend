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

// PATCH /api/me
func UpdateProfile(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid userId"})
	}
	update := bson.M{}
	if input.Name != "" {
		update["name"] = input.Name
	}
	if input.Email != "" {
		update["email"] = strings.ToLower(input.Email)
	}
	if len(update) == 0 {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "No data to update"})
	}
	_, err = userCol.UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": update})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to update profile"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Profile updated"})
}

// PATCH /api/me/password
func UpdatePassword(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}
	var input struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	objId, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid userId"})
	}
	var user models.User
	err = userCol.FindOne(ctx, bson.M{"_id": objId}).Decode(&user)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "User not found"})
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword)) != nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Current password is incorrect"})
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	_, err = userCol.UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": bson.M{"password": string(hash)}})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to update password"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Password updated"})
}

// GET /api/me/summary
func MeSummary(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}
	objId, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid userId"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ambil semua deteksi user
	cursor, err := detectionCol.Find(ctx, bson.M{"userId": objId})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to fetch detections"})
	}
	var detections []models.Detection
	if err := cursor.All(ctx, &detections); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to decode detections"})
	}

	total := len(detections)
	if total == 0 {
		return c.JSON(fiber.Map{
			"success": true,
			"summary": fiber.Map{
				"total":          0,
				"averageEmotion": nil,
				"lastDetection":  nil,
				"recent":         []interface{}{},
			},
		})
	}

	// Hitung rata-rata emosi
	sum := map[string]float64{
		"neutral": 0, "happy": 0, "sad": 0, "angry": 0, "surprised": 0, "disgusted": 0,
	}
	for _, d := range detections {
		sum["neutral"] += d.Emotions.Neutral
		sum["happy"] += d.Emotions.Happy
		sum["sad"] += d.Emotions.Sad
		sum["angry"] += d.Emotions.Angry
		sum["surprised"] += d.Emotions.Surprised
		sum["disgusted"] += d.Emotions.Disgusted
	}
	avg := map[string]float64{}
	for k, v := range sum {
		avg[k] = v / float64(total)
	}

	// Deteksi terakhir
	last := detections[total-1]

	// 5 deteksi terakhir (reverse order)
	recent := []fiber.Map{}
	for i := total - 1; i >= 0 && len(recent) < 5; i-- {
		d := detections[i]
		dom := "neutral"
		max := d.Emotions.Neutral
		if d.Emotions.Happy > max {
			dom, max = "happy", d.Emotions.Happy
		}
		if d.Emotions.Sad > max {
			dom, max = "sad", d.Emotions.Sad
		}
		if d.Emotions.Angry > max {
			dom, max = "angry", d.Emotions.Angry
		}
		if d.Emotions.Surprised > max {
			dom, max = "surprised", d.Emotions.Surprised
		}
		if d.Emotions.Disgusted > max {
			dom, max = "disgusted", d.Emotions.Disgusted
		}
		recent = append(recent, fiber.Map{
			"timestamp":   d.Timestamp,
			"dominant":    dom,
			"probability": max,
		})
	}

	// Hitung histogram dominan emosi
	histogram := map[string]int{
		"neutral": 0, "happy": 0, "sad": 0, "angry": 0, "surprised": 0, "disgusted": 0,
	}
	for _, d := range detections {
		dom := "neutral"
		max := d.Emotions.Neutral
		if d.Emotions.Happy > max {
			dom, max = "happy", d.Emotions.Happy
		}
		if d.Emotions.Sad > max {
			dom, max = "sad", d.Emotions.Sad
		}
		if d.Emotions.Angry > max {
			dom, max = "angry", d.Emotions.Angry
		}
		if d.Emotions.Surprised > max {
			dom, max = "surprised", d.Emotions.Surprised
		}
		if d.Emotions.Disgusted > max {
			dom, max = "disgusted", d.Emotions.Disgusted
		}
		histogram[dom]++
	}

	// Dominan rata-rata
	dominant := "neutral"
	maxAvg := avg["neutral"]
	for k, v := range avg {
		if v > maxAvg {
			dominant = k
			maxAvg = v
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"summary": fiber.Map{
			"total": total,
			"averageEmotion": fiber.Map{
				"distribution": avg,
				"dominant":     dominant,
			},
			"histogram": histogram,
			"lastDetection": fiber.Map{
				"timestamp": last.Timestamp,
				"emotions":  last.Emotions,
			},
			"recent": recent,
		},
	})
}
