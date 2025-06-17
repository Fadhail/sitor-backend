package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CameraStatus struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID   primitive.ObjectID `bson:"groupId" json:"groupId"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	IsActive  bool               `bson:"isActive" json:"isActive"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
