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
)

// POST /api/groups/:groupId/end-session
func EndSession(c *fiber.Ctx) error {
	groupId := c.Params("groupId")
	fmt.Println("[END-SESSION] groupId:", groupId)
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

	// Hapus semua status kamera user di grup ini (disconnect semua user dari sesi)
	res1, err := db.Collection("camera_status").DeleteMany(ctx, bson.M{"groupId": objGroupId})
	fmt.Println("[END-SESSION] camera_status deleted:", res1.DeletedCount, "error:", err)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to end session", "error": err.Error()})
	}

	// Set sessionActive=false pada group
	res2, err := db.Collection("groups").UpdateOne(ctx, bson.M{"_id": objGroupId}, bson.M{"$set": bson.M{"sessionActive": false}})
	fmt.Println("[END-SESSION] group update matched:", res2.MatchedCount, "modified:", res2.ModifiedCount, "error:", err)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to update group session status", "error": err.Error()})
	}

	// --- ARSIPKAN DETEKSI EMOSI SAAT END SESSION ---
	// Ambil semua deteksi emosi aktif untuk grup ini
	detectionsCursor, err := db.Collection("detections").Find(ctx, bson.M{"groupId": objGroupId})
	if err != nil {
		fmt.Println("[END-SESSION] gagal ambil deteksi:", err)
	} else {
		var detections []bson.M
		err = detectionsCursor.All(ctx, &detections)
		if err == nil && len(detections) > 0 {
			// Simpan ke koleksi detection_history
			historyDoc := bson.M{
				"groupId":    objGroupId,
				"sessionId":  primitive.NewObjectID(),
				"detections": detections,
				"endedAt":    time.Now(),
			}
			_, err = db.Collection("detection_history").InsertOne(ctx, historyDoc)
			if err != nil {
				fmt.Println("[END-SESSION] gagal simpan ke detection_history:", err)
			}
		}
	}
	// Hapus deteksi emosi aktif dari koleksi utama
	_, _ = db.Collection("detections").DeleteMany(ctx, bson.M{"groupId": objGroupId})

	return c.JSON(fiber.Map{"success": true, "message": "Sesi grup berhasil diakhiri. Semua user disconnect."})
}

// POST /api/groups/:groupId/start-session
func StartSession(c *fiber.Ctx) error {
	groupId := c.Params("groupId")
	fmt.Println("[START-SESSION] groupId:", groupId)
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

	// Set sessionActive=true pada group
	res, err := db.Collection("groups").UpdateOne(ctx, bson.M{"_id": objGroupId}, bson.M{"$set": bson.M{"sessionActive": true}})
	fmt.Println("[START-SESSION] group update matched:", res.MatchedCount, "modified:", res.ModifiedCount, "error:", err)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to update group sessionActive", "error": err.Error()})
	}

	// Tambahkan log untuk memastikan update berhasil
	var group models.Group
	err = db.Collection("groups").FindOne(ctx, bson.M{"_id": objGroupId}).Decode(&group)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to fetch group after update", "error": err.Error()})
	}
	fmt.Println("[START-SESSION] sessionActive after update:", group.SessionActive)

	// Ambil anggota grup
	if len(group.Members) > 0 {
		// Hapus status kamera lama (jika ada)
		_, _ = db.Collection("camera_status").DeleteMany(ctx, bson.M{"groupId": objGroupId})
		// Inisialisasi status kamera semua anggota ke isActive: false
		for _, userId := range group.Members {
			db.Collection("camera_status").InsertOne(ctx, bson.M{
				"groupId":   objGroupId,
				"userId":    userId,
				"isActive":  false,
				"updatedAt": time.Now(),
			})
		}
	}

	// (Opsional) Buat sesi baru di koleksi deteksi jika ingin tracking per sesi
	//
	// Bisa generate sessionId baru di sini jika ingin, lalu frontend kirim sessionId ke deteksi emosi

	return c.JSON(fiber.Map{"success": true, "message": "Sesi baru berhasil dimulai."})
}
