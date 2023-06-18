package main

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
	CreatedAt     string             `bson:"createdat"`
	UpdatedAt     string             `bson:"updatedat"`
	DeactivatedAt string             `bson:"deactivatedat"`
}

type BusinessModel struct {
	ObjectId     primitive.ObjectID `bson:"_id"`
	Name         string             `bson:"name"`
	Flatname     string             `bson:"flatname"`
	Description  string             `bson:"description"`
	Rating       float64            `bson:"rating"`
	RatingsCount int64              `bson:"ratingscount"`
	LatestReview string             `bson:"latestreview"`
	Companies    []string           `bson:"companies"`
	CreatedAt    string             `bson:"createdat"`
	UpdatedAt    string             `bson:"updatedat"`
}
