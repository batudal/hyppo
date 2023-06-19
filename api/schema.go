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
	CreatedAt     int64              `bson:"createdat"`
	UpdatedAt     int64              `bson:"updatedat"`
	DeactivatedAt int64              `bson:"deactivatedat"`
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
	CreatedAt    int64              `bson:"createdat"`
	UpdatedAt    int64              `bson:"updatedat"`
}

type Feed struct {
	Page   int64           `bson:"page"`
	SortBy string          `bson:"sortby"`
	Models []BusinessModel `bson:"models"`
}
