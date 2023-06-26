package handlers

import (
	"context"
	"os"
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
		page_size := os.Getenv("APP_MODEL_PAGE_SIZE")
		page_size_int, err := strconv.Atoi(page_size)
		if err != nil {
			return err
		}
		filter := bson.D{}
		opts := options.
			Find().
			SetSort(bson.D{{c.Query("sortby"), order}}).
			SetLimit(int64(page_size_int)).
			SetSkip(int64(page-1) * int64(page_size_int))
		coll := cfg.Mc.Database("primary").Collection("business-models")
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err = cursor.All(context.TODO(), &feed.Models); err != nil {
			panic(err)
		}
		return c.Render("partials/model/paged-models", fiber.Map{
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
			page_size := os.Getenv("APP_MODEL_PAGE_SIZE")
			page_size_int, err := strconv.Atoi(page_size)
			if err != nil {
				return err
			}
			filter := bson.D{}
			opts := options.
				Find().
				SetSort(bson.D{{"createdat", -1}}).
				SetLimit(int64(page_size_int))
			coll := cfg.Mc.Database("primary").Collection("business-models")
			cursor, err := coll.Find(context.Background(), filter, opts)
			if err = cursor.All(context.TODO(), &feed.Models); err != nil {
				panic(err)
			}
			return c.Render("partials/model/paged-models", fiber.Map{
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
		return c.Render("partials/model/search-models", fiber.Map{
			"Models": models,
		})
	}
}
