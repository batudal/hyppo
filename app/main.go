package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/mongodb"
	"github.com/gofiber/template/html/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v74"
	"github.com/trycourier/courier-go/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	mc      *mongo.Client
	store   *session.Store
	redis   *redis.Client
	courier *courier.Client
}

func main() {
	storage := mongodb.New(mongodb.Config{
		ConnectionURI: os.Getenv("MONGODB_URI"),
		Database:      "secondary",
		Collection:    "sessions",
		Reset:         false,
	})
	store := session.New(session.Config{
		Storage: storage,
	})
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	client, err := NewMongoClient()
	if err != nil {
		panic(err)
	}
	engine := html.New("./views", ".html")
	if os.Getenv("PRODUCTION") == "0" {
		engine.Reload(true)
	}
	redis_client := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		DB:   0,
	})
	courier_client := courier.CreateClient(os.Getenv("COURIER_TOKEN"), nil)
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	cfg := Config{
		mc:      client,
		store:   store,
		redis:   redis_client,
		courier: courier_client,
	}
	app.Use(logger.New())
	// Pages
	app.Get("/", IndexPage(cfg))
	app.Get("/welcome", HandleWelcomePage(cfg)) // stripe callback
	// Webhooks
	app.Post("/hooks/subscribe", HandleCreateMember(cfg)) // stripe webhook
	app.Post("/hooks/cancel", HandleCancelMember(cfg))    // stripe webhook
	// Partials
	app.Get("/models", HandleGetModels(cfg))
	app.Post("/login", HandleLogin(cfg))
	app.Get("/review/new", Authorize(cfg), NewReview(cfg))
	// Redirects
	app.Get("/logout", HandleLogout(cfg))
	app.Get("/magic/:email/:magic", HandleMagic(cfg))
	// Modals
	app.Get("/modals/welcome", HandleWelcomeModal(cfg))
	app.Get("/modals/reviews", Authorize(cfg), HandleReviewsModal(cfg))
	// Search
	app.Post("/search", HandleSearch(cfg))
	// Actions
	app.Post("/review", Authorize(cfg), HandleComment(cfg))
	app.Static("/assets", "./assets")
	app.Listen(":80")
}
