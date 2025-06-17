package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatMessage struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	GroupID   primitive.ObjectID  `bson:"groupId" json:"groupId"`
	UserID    *primitive.ObjectID `bson:"userId,omitempty" json:"userId"`
	UserName  string              `bson:"userName" json:"userName"`
	Role      string              `bson:"role" json:"role"`
	Message   string              `bson:"message" json:"message"`
	Timestamp time.Time           `bson:"timestamp" json:"timestamp"`
	IsFromAI  bool                `bson:"isFromAI" json:"isFromAI"`
}
