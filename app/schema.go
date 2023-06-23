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
	LatestReview string             `bson:"latestreview"`
	ReviewCount  int64              `bson:"reviewcount"`
	LatestTest   string             `bson:"latesttest"`
	TestCount    int64              `bson:"testcount"`
	Companies    []string           `bson:"companies"`
	CreatedAt    int64              `bson:"createdat"`
	UpdatedAt    int64              `bson:"updatedat"`
}

type Review struct {
	ObjectId     primitive.ObjectID `bson:"_id"`
	ModelId      primitive.ObjectID `bson:"modelid"`
	UserId       primitive.ObjectID `bson:"userid"`
	Comment      string             `bson:"comment"`
	CreatedAt    int64              `bson:"createdat"`
	UpdatedAt    int64              `bson:"updatedat"`
	HelpfulCount int64              `bson:"helpfulcount"`
}

type Feed struct {
	Page   int64           `bson:"page"`
	SortBy string          `bson:"sortby"`
	Models []BusinessModel `bson:"models"`
}
