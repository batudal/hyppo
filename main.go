package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v74"
	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	mc *mongo.Client
}

func main() {
	godotenv.Load()
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	client, err := NewMongoClient()
	if err != nil {
		panic(err)
	}
	engine := html.New("./views", ".html")
	if os.Getenv("PRODUCTION") == "0" {
		engine.Reload(true)
	}
	engine.Reload(true)
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	config := Config{
		mc: client,
	}
	app.Use(logger.New())
	app.Get("/", IndexPage(config.mc))
	app.Get("/create_user", HandleCreateUser(config.mc))
	app.Get("/models", HandleGetModels(config.mc))
	app.Static("/assets", "./assets")
	app.Listen(":80")
}
