package controllers

import (
	"context"
	"fmt"
	"time"

	"sitor-backend/config"
	"sitor-backend/models"
	"sitor-backend/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var groupCol = config.GetDB().Collection("groups")

// GET /api/groups
func GetGroups(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := groupCol.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println("[ERROR] groupCol.Find:", err)
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to fetch groups", "error": err.Error()})
	}
	var groups []models.Group
	if err := cursor.All(ctx, &groups); err != nil {
		fmt.Println("[ERROR] cursor.All:", err)
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to decode groups", "error": err.Error()})
	}
	var groupsWithStringMembers []map[string]interface{}
	for _, g := range groups {
		// Fallback jika LeaderID/Members nil
		leaderIdStr := ""
		if g.LeaderID != primitive.NilObjectID {
			leaderIdStr = g.LeaderID.Hex()
		}
		members := make([]string, 0)
		if g.Members != nil {
			for i, m := range g.Members {
				if m != primitive.NilObjectID {
					members = append(members, m.Hex())
				} else {
					fmt.Println("[WARN] NilObjectID in members at index", i)
				}
			}
		}
		groupsWithStringMembers = append(groupsWithStringMembers, map[string]interface{}{
			"id":          g.ID.Hex(),
			"name":        g.Name,
			"description": g.Description,
			"leaderId":    leaderIdStr,
			"members":     members,
			"createdAt":   g.CreatedAt,
		})
	}
	return c.JSON(fiber.Map{"success": true, "groups": groupsWithStringMembers})
}

func CreateGroup(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}
	type reqBody struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		SecurityCode string `json:"securityCode"`
	}
	var body reqBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid request"})
	}
	hash, _ := utils.HashPassword(body.SecurityCode)
	leaderObjId, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid userId"})
	}
	group := models.Group{
		ID:            primitive.NewObjectID(),
		Name:          body.Name,
		Description:   body.Description,
		SecurityCode:  hash,
		LeaderID:      leaderObjId,
		Members:       []primitive.ObjectID{leaderObjId},
		CreatedAt:     time.Now(),
		SessionActive: true, // Pastikan sesi aktif saat grup dibuat
	}
	_, err = groupCol.InsertOne(context.TODO(), group)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to create group"})
	}
	return c.JSON(fiber.Map{"success": true, "group": group})
}

// POST /api/groups/join
func JoinGroup(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}
	var body models.JoinGroupRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid request"})
	}
	objGroupId, err := primitive.ObjectIDFromHex(body.GroupId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid groupId"})
	}
	var group models.Group
	err = groupCol.FindOne(context.TODO(), bson.M{"_id": objGroupId}).Decode(&group)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Group not found"})
	}
	if !utils.CheckPasswordHash(body.SecurityCode, group.SecurityCode) {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Invalid security code"})
	}
	userObjId, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid userId"})
	}
	for _, m := range group.Members {
		if m.Hex() == userObjId.Hex() {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Already joined"})
		}
	}
	_, err = groupCol.UpdateOne(context.TODO(), bson.M{"_id": objGroupId}, bson.M{"$push": bson.M{"members": userObjId}})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to join group"})
	}
	return c.JSON(fiber.Map{"success": true})
}

func ListGroupMembers(c *fiber.Ctx) error {
	groupId := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid groupId"})
	}
	var group models.Group
	err = groupCol.FindOne(context.TODO(), bson.M{"_id": objId}).Decode(&group)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Group not found"})
	}
	return c.JSON(fiber.Map{"success": true, "members": group.Members})
}

// DELETE /api/groups/:id
func DeleteGroup(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}
	groupId := c.Params("id")
	objGroupId, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid groupId"})
	}
	var group models.Group
	err = groupCol.FindOne(context.TODO(), bson.M{"_id": objGroupId}).Decode(&group)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Group not found"})
	}
	if group.LeaderID.Hex() != userId.(string) {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Only the group leader can delete the group"})
	}
	_, err = groupCol.DeleteOne(context.TODO(), bson.M{"_id": objGroupId})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to delete group"})
	}
	return c.JSON(fiber.Map{"success": true})
}

// POST /api/groups/:id/leave
func LeaveGroup(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}
	groupId := c.Params("id")
	objGroupId, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid groupId"})
	}
	var group models.Group
	err = groupCol.FindOne(context.TODO(), bson.M{"_id": objGroupId}).Decode(&group)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Group not found"})
	}
	if group.LeaderID.Hex() == userId.(string) {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Leader cannot leave the group. Please delete the group instead."})
	}
	userObjId, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid userId"})
	}
	_, err = groupCol.UpdateOne(context.TODO(), bson.M{"_id": objGroupId}, bson.M{"$pull": bson.M{"members": userObjId}})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to leave group"})
	}
	return c.JSON(fiber.Map{"success": true})
}
