package main

import (
	"os"
	"time"

	"context"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/mongodb"
	"github.com/gofiber/template/html/v2"
	"github.com/stripe/stripe-go/v74"
	"github.com/trycourier/courier-go/v2"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg, app := setup()
	app.Use(logger.New())
	Routes(app, cfg)
}

func setup() (config.Config, *fiber.App) {
	redis_client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	storage := mongodb.New(mongodb.Config{
		ConnectionURI: os.Getenv("MONGODB_URI"),
		Database:      "secondary",
		Collection:    "sessions",
		Reset:         false,
	})
	store := session.New(session.Config{
		Storage:        storage,
		Expiration:     30 * 24 * time.Hour,
		CookieDomain:   os.Getenv("COOKIE_DOMAIN"),
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: "Strict",
	})
	store.RegisterType(&schema.User{})
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	uri := os.Getenv("MONGODB_URI")
	mongodb_client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	engine := html.New("./views", ".html")
	if os.Getenv("PRODUCTION") == "0" {
		engine.Reload(true)
	}
	courier_client := courier.CreateClient(os.Getenv("COURIER_TOKEN"), nil)
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	cfg := config.Config{
		Mc:      mongodb_client,
		Store:   store,
		Courier: courier_client,
		Redis:   redis_client,
	}
	return cfg, app
}
