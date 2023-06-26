package schema

import (
	"os"
	"strconv"
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

type ReviewFeed struct {
	Reviews []Review
	Page    int64  `bson:"page"`
	SortBy  string `bson:"sortby"`
}

func (r Review) ParseDate() string {
	date := time.Unix(r.UpdatedAt, 0)
	return date.Format("January 2, 2006")
}

func (m Model) IsLastReview(i int) bool {
	page_size := os.Getenv("APP_REVIEW_PAGE_SIZE")
	page_size_int, err := strconv.Atoi(page_size)
	if err != nil {
		return false
	}
	return i == page_size_int
}
