package handlers

import (
	"context"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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