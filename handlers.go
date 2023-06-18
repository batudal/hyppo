package main

import (
	"context"
	"html/template"
	"strconv"

	// "html/template"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/stripe/stripe-go/v74/webhook"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func HandleGetModels(db *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			return err
		}
		feed := Feed{
			Page:   int64(page),
			SortBy: c.Query("sortby"),
		}
		filter := bson.D{}
		opts := options.
			Find().
			SetSort(bson.D{{c.Query("sortby"), -1}}).
			SetLimit(4).
			SetSkip(int64(page-1) * 4)
		coll := db.Database("primary").Collection("business-models")
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err = cursor.All(context.TODO(), &feed.Models); err != nil {
			panic(err)
		}
		return c.Render("partials/business-model", fiber.Map{
			"Feed": feed,
		})
	}
}

func IndexPage(db *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		feed := Feed{
			Page:   1,
			SortBy: "createdat",
		}
		filter := bson.D{}
		opts := options.Find().SetSort(bson.D{{"createdat", -1}}).SetLimit(4)
		coll := db.Database("primary").Collection("business-models")
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err = cursor.All(context.TODO(), &feed.Models); err != nil {
			panic(err)
			// todo: return 404 page
		}
		return c.Render("pages/index", fiber.Map{
			"Feed": feed,
		}, "layouts/main")
	}
}

func (m BusinessModel) IsLast(i int) bool {
	return i == 3
}

func (m BusinessModel) Increment(i int64) int64 {
	return i + 1
}

func (m BusinessModel) ParseDescription() template.HTML {
	buf := mdToHTML([]byte(m.Description))
	return template.HTML(buf)
}

func mdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return markdown.Render(doc, renderer)
}

func HandleCreateUser(db *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		event, err := webhook.ConstructEvent(c.Body(), c.GetReqHeaders()["Stripe-Signature"], os.Getenv("STRIPE_CREATEUSER_SECRET"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Error verifying webhook signature")
		}
		switch event.Type {
		case "customer.created":
			// todo: send otp email
			user := User{
				Email: event.Data.Object["email"].(string),
				Name:  event.Data.Object["name"].(string),
			}
			coll := db.Database("primary").Collection("users")
			_, err = coll.InsertOne(context.Background(), user)
			if err != nil {
				return err
			}
		default:
		}
		return c.SendStatus(fiber.StatusOK)
	}
}
