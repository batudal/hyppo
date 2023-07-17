package main

import (
	"flag"
	"os"
	"time"

	"context"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/mongodb"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v74"
	"github.com/trycourier/courier-go/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg, app := setup()
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowHeaders:  "HX-Request, HX-Trigger, HX-Trigger-Name, HX-Target, HX-Prompt",
		ExposeHeaders: "HX-Push, HX-Redirect, HX-Location, HX-Refresh, HX-Trigger, HX-Trigger-After-Swap, HX-Trigger-After-Settle",
	}))
	Routes(app, cfg)
}

func setup() (config.Config, *fiber.App) {
	dev := flag.Bool("dev", false, "development mode")
	flag.Parse()
	if *dev {
		err := godotenv.Load("../.env")
		if err != nil {
			panic(err)
		}
	}
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
	if *dev {
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
