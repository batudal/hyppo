package handlers

import (
	"context"
	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/batudal/hyppo/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
		var reviews []schema.Review
		cursor, err := coll.Find(context.Background(), bson.M{"modelid": model_id})
		if err != nil {
			return err
		}
		if err = cursor.All(context.Background(), &reviews); err != nil {
			return err
		}
		var user *schema.User
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		if sess.Get("user") != nil {
			user = sess.Get("user").(*schema.User)
		} else {
			user = &schema.User{}
		}
		authored_reviews := utils.GetAuthoredReviews(cfg, reviews, *user)
		return c.Render("pages/model/reviews", fiber.Map{
			"User":    user,
			"Model":   model,
			"Reviews": authored_reviews,
		})
	}
}
