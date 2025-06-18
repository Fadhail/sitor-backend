package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Emotion struct {
	Neutral   float64 `bson:"neutral" json:"neutral"`
	Happy     float64 `bson:"happy" json:"happy"`
	Sad       float64 `bson:"sad" json:"sad"`
	Angry     float64 `bson:"angry" json:"angry"`
	Surprised float64 `bson:"surprised" json:"surprised"`
	Disgusted float64 `bson:"disgusted" json:"disgusted"`
}

type Detection struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID   primitive.ObjectID `bson:"groupId" json:"groupId"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	UserName  string             `bson:"userName" json:"userName"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	Date      string             `bson:"date" json:"date"`
	Emotions  Emotion            `bson:"emotions" json:"emotions"`
}
