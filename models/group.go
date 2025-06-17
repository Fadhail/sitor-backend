package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name         string               `bson:"name" json:"name"`
	Description  string               `bson:"description" json:"description"`
	SecurityCode string               `bson:"securityCode" json:"securityCode"`
	LeaderID     primitive.ObjectID   `bson:"leaderId" json:"leaderId"`
	Members      []primitive.ObjectID `bson:"members" json:"members"`
	CreatedAt    time.Time            `bson:"createdAt" json:"createdAt"`
}

// Request struct khusus untuk join group
// Digunakan pada handler JoinGroup di controllers
// Hanya menerima groupId dan securityCode dari frontend
type JoinGroupRequest struct {
	GroupId      string `json:"groupId"`
	SecurityCode string `json:"securityCode"`
}
