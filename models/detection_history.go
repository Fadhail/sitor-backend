package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DetectionHistory struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID    primitive.ObjectID `bson:"groupId" json:"groupId"`
	SessionID  primitive.ObjectID `bson:"sessionId" json:"sessionId"`
	Detections []Detection        `bson:"detections" json:"detections"`
	StartedAt  time.Time          `bson:"startedAt" json:"startedAt"`
	EndedAt    time.Time          `bson:"endedAt" json:"endedAt"`
}
