package handlers

import (
	"context"
	"strconv"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func HandleGetModels(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			return err
		}
		order, err := strconv.Atoi(c.Query("order"))
		if err != nil {
			return err
		}
		feed := schema.Feed{
			Page:   int64(page),
			SortBy: c.Query("sortby"),
		}
		filter := bson.D{}
		opts := options.
			Find().
			SetSort(bson.D{{c.Query("sortby"), order}}).
			SetLimit(4).
			SetSkip(int64(page-1) * 4)
		coll := cfg.Mc.Database("primary").Collection("business-models")
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err = cursor.All(context.TODO(), &feed.Models); err != nil {
			panic(err)
		}
		return c.Render("partials/business-model", fiber.Map{
			"Feed": feed,
		})
	}
}

func HandleSearch(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.FormValue("query") == "" {
			feed := schema.Feed{
				Page:   1,
				SortBy: "createdat",
			}
			filter := bson.D{}
			opts := options.
				Find().
				SetSort(bson.D{{"createdat", -1}}).
				SetLimit(4)
			coll := cfg.Mc.Database("primary").Collection("business-models")
			cursor, err := coll.Find(context.Background(), filter, opts)
			if err = cursor.All(context.TODO(), &feed.Models); err != nil {
				panic(err)
			}
			return c.Render("partials/business-model", fiber.Map{
				"Feed": feed,
			})
		}
		coll := cfg.Mc.Database("primary").Collection("business-models")
		filter := bson.D{{"$text", bson.D{{"$search", c.FormValue("query")}}}}
		var models []schema.Model
		cursor, err := coll.Find(context.TODO(), filter)
		if err != nil {
			return err
		}
		if err = cursor.All(context.Background(), &models); err != nil {
			return err
		}
		return c.Render("partials/business-models", fiber.Map{
			"Models": models,
		})
	}
}
