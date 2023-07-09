package utils

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/trycourier/courier-go/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type HumanDate struct {
	StartDate string
	EndDate   string
}

func HumanizeDates(tests *[]schema.Test) ([]HumanDate, error) {
	var human_dates []HumanDate
	for _, test := range *tests {
		startdate := test.StartDate
		startdate_human := startdate.Time().Format("January 2, 2006")
		enddate := test.EndDate
		enddate_human := enddate.Time().Format("January 2, 2006")
		human_dates = append(human_dates, HumanDate{startdate_human, enddate_human})
	}
	return human_dates, nil
}

func UpdateTestCounts(
	cfg config.Config,
	old primitive.ObjectID,
	new primitive.ObjectID) error {
	if old == new {
		return nil
	}
	if old != primitive.NilObjectID {
		err := DecrementTestCount(cfg, old)
		if err != nil {
			return err
		}
	}
	err := IncrementTestCount(cfg, new)
	if err != nil {
		return err
	}
	return nil
}

func DecrementTestCount(cfg config.Config, id primitive.ObjectID) error {
	coll := cfg.Mc.Database("primary").Collection("business-models")
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$inc", bson.D{{"testcount", -1}}}}
	_, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func IncrementTestCount(cfg config.Config, id primitive.ObjectID) error {
	coll := cfg.Mc.Database("primary").Collection("business-models")
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$inc", bson.D{{"testcount", 1}}}}
	_, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func GetAllModels(cfg config.Config) ([]schema.Model, error) {
	coll := cfg.Mc.Database("primary").Collection("business-models")
	var models []schema.Model
	cursor, err := coll.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &models); err != nil {
		return nil, err
	}
	return models, nil
}

func GetAllMethods(cfg config.Config) ([]schema.Method, error) {
	coll := cfg.Mc.Database("primary").Collection("methods")
	var methods []schema.Method
	cursor, err := coll.Find(context.Background(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &methods); err != nil {
		return nil, err
	}
	return methods, nil
}

func GetAuthoredReviews(cfg config.Config, reviews []schema.Review, user schema.User) []schema.AuthoredReview {
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
			coll_helpfuls := cfg.Mc.Database("primary").Collection("helpfuls")
			filter = bson.D{{"reviewid", review.ObjectId}}
			count, err := coll_helpfuls.CountDocuments(context.Background(), filter)
			if err != nil {
				authored_reviews[i].HelpfulCount = 0
			} else {
				authored_reviews[i].HelpfulCount = count
			}
			if !user.ObjectId.IsZero() {
				filter = bson.D{{"reviewid", review.ObjectId}, {"userid", user.ObjectId}}
				count, err = coll_helpfuls.CountDocuments(context.Background(), filter)
				if err != nil {
					authored_reviews[i].Helpful = false
				} else {
					authored_reviews[i].Helpful = count > 0
				}
			}
		}(i, review, authored_reviews, &wg)
	}
	wg.Wait()
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
