package schema

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ObjectId      primitive.ObjectID `bson:"_id"`
	Name          string             `bson:"name"`
	Avatar        string             `bson:"avatar"`
	Company       string             `bson:"company"`
	Email         string             `bson:"email"`
	Website       string             `bson:"website"`
	Socials       []string           `bson:"socials"`
	Membership    bool               `bson:"membership"`
	MembershipAt  int64              `bson:"membershipat"`
	MembershipEnd int64              `bson:"membershipend"`
	StripeId      string             `bson:"stripeid"`
	CreatedAt     int64              `bson:"createdat"`
	UpdatedAt     int64              `bson:"updatedat"`
	DeactivatedAt int64              `bson:"deactivatedat"`
}
