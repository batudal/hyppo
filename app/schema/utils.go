package schema

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthoredReview struct {
	Review       Review
	Author       User
	Helpful      bool
	HelpfulCount int64
}

type Helpful struct {
	UserId   primitive.ObjectID `bson:"userid"`
	ReviewId primitive.ObjectID `bson:"reviewid"`
}
