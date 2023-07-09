package schema

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Method struct {
	ObjectId    primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	Flatname    string             `bson:"flatname"`
	Description string             `bson:"description"`
	CreatedAt   primitive.DateTime `bson:"createdat"`
	UpdatedAt   primitive.DateTime `bson:"updatedat"`
}
