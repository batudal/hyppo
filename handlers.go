package main

import (
	"context"
	"html/template"
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

func IndexPage(db *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var models []BusinessModel
		filter := bson.D{}
		opts := options.Find().SetSort(bson.D{{"rating", -1}})
		coll := db.Database("primary").Collection("business-models")
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err = cursor.All(context.TODO(), &models); err != nil {
			panic(err)
			// todo: return 404 page
		}
		return c.Render("pages/index", fiber.Map{
			"Models": models,
		}, "layouts/main")
	}
}

func (m BusinessModel) ParseDescription() template.HTML {
	buf := mdToHTML([]byte(m.Description))
	return template.HTML(buf)
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)
	// create HTML renderer with extensions
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
