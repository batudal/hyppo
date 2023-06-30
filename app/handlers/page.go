package handlers

import (
	"context"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ModelPage(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		flatname := c.Params("flatname")
		coll := cfg.Mc.Database("primary").Collection("business-models")
		filter := bson.D{{"flatname", flatname}}
		var model schema.Model
		err := coll.FindOne(context.Background(), filter).Decode(&model)
		if err != nil {
			return err
		}
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user")
		return c.Render("pages/model", fiber.Map{
			"User":  user,
			"Model": model,
		}, "layouts/page")
	}
}

func IndexPage(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			panic(err)
		}
		user := sess.Get("user")
		feed := schema.Feed{
			Page:   1,
			SortBy: "createdat",
		}
		filter := bson.D{}
		opts := options.Find().SetSort(bson.D{{"createdat", -1}}).SetLimit(4)
		coll := cfg.Mc.Database("primary").Collection("business-models")
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err = cursor.All(context.TODO(), &feed.Models); err != nil {
			return err
		}
		return c.Render("pages/index", fiber.Map{
			"View": "feed",
			"User": user,
			"Feed": feed,
		}, "layouts/main")
	}
}

func HandleWelcomePage(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("pages/index", fiber.Map{
			"Modal": "welcome",
		}, "layouts/main")
	}
}
