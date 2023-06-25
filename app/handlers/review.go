package handlers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewReview(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(schema.User)
		if !user.Membership {
			return fiber.NewError(fiber.StatusForbidden, "You must be a member to review models.")
		}
		model_id, err := primitive.ObjectIDFromHex(c.Query("model_id"))
		if err != nil {
			return err
		}
		fmt.Println("model_id: ", model_id)
		var model schema.Model
		models_coll := cfg.Mc.Database("primary").Collection("business-models")
		filter := bson.D{{"_id", model_id}}
		err = models_coll.FindOne(context.Background(), filter).Decode(&model)
		if err != nil {
			return err
		}
		return c.Render("partials/new_review", fiber.Map{
			"Model": model,
			"User":  user,
		})
	}
}

func HandleReviewsModal(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type AuthoredReview struct {
			Review schema.Review
			Author schema.User
		}
		coll := cfg.Mc.Database("primary").Collection("reviews")
		model_id, err := primitive.ObjectIDFromHex(c.Query("model_id"))
		if err != nil {
			return err
		}
		filter := bson.D{{"modelid", model_id}}
		opts := options.Find().SetSort(bson.D{{"createdat", -1}}).SetLimit(4).SetSkip(0)
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err != nil {
			return err
		}
		var reviews []schema.Review
		if err = cursor.All(context.Background(), &reviews); err != nil {
			return err
		}
		wg := sync.WaitGroup{}
		authored_reviews := make([]AuthoredReview, len(reviews))
		for i, review := range reviews {
			wg.Add(1)
			go func(i int, review schema.Review, authored_reviews []AuthoredReview, wg *sync.WaitGroup) {
				authored_reviews[i].Review = review
				fmt.Println(i, " ", review)
				filter = bson.D{{"_id", review.UserId}}
				err = cfg.Mc.Database("primary").Collection("users").FindOne(context.Background(), filter).Decode(&authored_reviews[i].Author)
				if err != nil {
					authored_reviews[i].Author = schema.User{
						Name: "Deleted User",
					}
				}
				wg.Done()
			}(i, review, authored_reviews, &wg)
		}
		wg.Wait()
		coll = cfg.Mc.Database("primary").Collection("business-models")
		filter = bson.D{{"_id", model_id}}
		var model schema.Model
		err = coll.FindOne(context.Background(), filter).Decode(&model)
		if err != nil {
			return err
		}
		return c.Render("modals/reviews", fiber.Map{
			"Model":   model,
			"Reviews": authored_reviews,
		})
	}
}

func HandleComment(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.FormValue("comment") == "" {
			return c.Render("partials/comment", fiber.Map{
				"Error": "Comment cannot be empty",
			})
		}
		fmt.Println("model_id: ", c.FormValue("model_id"))
		model_id, err := primitive.ObjectIDFromHex(c.FormValue("model_id"))
		if err != nil {
			return err
		}
		user_id := c.Locals("user_id").(primitive.ObjectID)
		var user schema.User
		filter := bson.D{{"_id", user_id}}
		err = cfg.Mc.Database("primary").Collection("users").FindOne(context.Background(), filter).Decode(&user)
		if err != nil {
			return err
		}
		review := schema.Review{
			ObjectId:  primitive.NewObjectID(),
			ModelId:   model_id,
			UserId:    user_id,
			Comment:   c.FormValue("comment"),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		reviews_coll := cfg.Mc.Database("primary").Collection("reviews")
		_, err = reviews_coll.InsertOne(context.Background(), review)
		if err != nil {
			return err
		}
		models_coll := cfg.Mc.Database("primary").Collection("business-models")
		filter = bson.D{{"_id", model_id}}
		update := bson.D{{"$set", bson.D{{"latestreview", review.Comment}}}}
		_, err = models_coll.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}
		update_review_count := bson.D{{"$inc", bson.D{{"reviewcount", 1}}}}
		_, err = models_coll.UpdateOne(context.Background(), filter, update_review_count)
		if err != nil {
			return err
		}
		return c.Render("partials/comment", fiber.Map{
			"User":   user,
			"Review": review,
		})
	}
}
