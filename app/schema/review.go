package schema

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Review struct {
	ObjectId     primitive.ObjectID `bson:"_id"`
	ModelId      primitive.ObjectID `bson:"modelid"`
	UserId       primitive.ObjectID `bson:"userid"`
	Comment      string             `bson:"comment"`
	CreatedAt    int64              `bson:"createdat"`
	UpdatedAt    int64              `bson:"updatedat"`
	HelpfulCount int64              `bson:"helpfulcount"`
}

func (r Review) ParseDate() string {
	date := time.Unix(r.UpdatedAt, 0)
	return date.Format("January 2, 2006")
}
