package utils

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/trycourier/courier-go/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAuthoredReviews(cfg config.Config, reviews []schema.Review) []schema.AuthoredReview {
	fmt.Println("Hello getting authored reviews")
	var wg sync.WaitGroup
	wg.Add(len(reviews))
	authored_reviews := make([]schema.AuthoredReview, len(reviews))
	for i, review := range reviews {
		go func(i int, review schema.Review, authored_reviews []schema.AuthoredReview, wg *sync.WaitGroup) {
			defer wg.Done()
			authored_reviews[i].Review = review
			filter := bson.D{{"_id", review.UserId}}
			err := cfg.Mc.
				Database("primary").
				Collection("users").
				FindOne(context.Background(), filter).
				Decode(&authored_reviews[i].Author)
			if err != nil {
				authored_reviews[i].Author = schema.User{
					Name: "Deleted User",
				}
			}
		}(i, review, authored_reviews, &wg)
	}
	wg.Wait()
	fmt.Println("Hello returning authored reviews")
	return authored_reviews
}

func CheckMembership(db *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user schema.User
		coll := db.Database("primary").Collection("users")
		err := coll.FindOne(context.Background(), c.Locals("user")).Decode(&user)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "You are not a member of this organization")
		}
		return c.Next()
	}
}

func RandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func SendEmail(client *courier.Client, to string, template string, data map[string]string) error {
	_, err := client.SendMessage(
		context.Background(),
		courier.SendMessageRequestBody{
			Message: map[string]interface{}{
				"to": map[string]string{
					"email": to,
				},
				"brand_id": "ZE1HM57RN74BBGPKNYQJ797YYZPN",
				"template": template,
				"data":     data,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func SaveMagic(client *redis.Client, mail_address string, magic string) error {
	key := "magic:" + mail_address
	err := client.Set(context.Background(), key, magic, 300*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func VerifyMagic(client *redis.Client, mail_address string, magic string) error {
	key := "magic:" + mail_address
	magic_db, err := client.Get(context.Background(), key).Result()
	if err != nil {
		return err
	}
	if magic_db != magic {
		return errors.New("")
	}
	return nil
}
