package handlers

import (
	"context"
	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/batudal/hyppo/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CompletedTab(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		coll := cfg.Mc.Database("primary").Collection("tests")
		filter := bson.D{{"userid", user.ObjectId}, {"state", "completed"}}
		var tests []schema.Test
		cursor, err := coll.Find(context.Background(), filter)
		if err != nil {
			return err
		}
		if err = cursor.All(context.Background(), &tests); err != nil {
			return err
		}
		human_dates, err := utils.HumanizeDates(&tests)
		if err != nil {
			return err
		}
		return c.Render("pages/tests/completed", fiber.Map{
			"User":       user,
			"Tests":      tests,
			"HumanDates": human_dates,
		})
	}
}

func OngoingTab(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		coll := cfg.Mc.Database("primary").Collection("tests")
		filter := bson.D{{"userid", user.ObjectId}, {"state", "ongoing"}}
		var tests []schema.Test
		cursor, err := coll.Find(context.Background(), filter)
		if err != nil {
			return err
		}
		if err = cursor.All(context.Background(), &tests); err != nil {
			return err
		}
		human_dates, err := utils.HumanizeDates(&tests)
		if err != nil {
			return err
		}
		return c.Render("pages/tests/ongoing", fiber.Map{
			"User":       user,
			"Tests":      tests,
			"HumanDates": human_dates,
		})
	}
}

func ModelTab(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		model_id, err := primitive.ObjectIDFromHex(c.Query("model_id"))
		if err != nil {
			return err
		}
		coll := cfg.Mc.Database("primary").Collection("business-models")
		var model schema.Model
		err = coll.FindOne(context.Background(), bson.M{"_id": model_id}).Decode(&model)
		if err != nil {
			return err
		}
		return c.Render("pages/model/model", fiber.Map{
			"Model": model,
		})
	}
}

func ReviewsTab(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		model_id, err := primitive.ObjectIDFromHex(c.Query("model_id"))
		if err != nil {
			return err
		}
		model_coll := cfg.Mc.Database("primary").Collection("business-models")
		var model schema.Model
		err = model_coll.FindOne(context.Background(), bson.M{"_id": model_id}).Decode(&model)
		if err != nil {
			return err
		}
		coll := cfg.Mc.Database("primary").Collection("reviews")
		opts := options.Find().SetSort(bson.D{{"createdat", -1}})
		var reviews []schema.Review
		cursor, err := coll.Find(context.Background(), bson.M{"modelid": model_id}, opts)
		if err != nil {
			return err
		}
		if err = cursor.All(context.Background(), &reviews); err != nil {
			return err
		}
		user := &schema.User{}
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		if sess.Get("user") != nil {
			user = sess.Get("user").(*schema.User)
			authored_reviews := utils.GetAuthoredReviews(cfg, reviews, *user)
			return c.Render("pages/model/reviews", fiber.Map{
				"User":    user,
				"Model":   model,
				"Reviews": authored_reviews,
			})
		} else {
			authored_reviews := utils.GetAuthoredReviews(cfg, reviews, *user)
			return c.Render("pages/model/reviews", fiber.Map{
				"Model":   model,
				"Reviews": authored_reviews,
			})
		}
	}
}
