package handlers

import (
	"context"
	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func HandleSubscribeModels(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Query("user") == "false" {
			coll := cfg.Mc.Database("primary").Collection("model-subscribers")
			_, err := coll.InsertOne(context.Background(), bson.D{{"email", c.Query("email")}})
			if err != nil {
				return err
			}
			return c.Status(fiber.StatusOK).Render("partials/subscribed", fiber.Map{
				"Subject": "Business Models",
				"Email":   c.Query("email"),
			})
		}
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		coll := cfg.Mc.Database("primary").Collection("users")
		filter := bson.D{{"_id", user.ObjectId}}
		update := bson.D{{"$set", bson.D{{"newsmodel", true}}}}
		_, err = coll.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}
		var updated_user schema.User
		err = coll.FindOne(context.Background(), filter).Decode(&updated_user)
		if err != nil {
			return err
		}
		sess.Set("user", &updated_user)
		err = sess.Save()
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).Render("partials/subscribed", fiber.Map{
			"Subject": "Business Models",
			"Email":   user.Email,
		})
	}
}

func HandleUnsubscribeModels(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		coll := cfg.Mc.Database("primary").Collection("users")
		filter := bson.D{{"_id", user.ObjectId}}
		update := bson.D{{"$set", bson.D{{"newsmodel", false}}}}
		_, err = coll.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).Render("", fiber.Map{})
	}
}
