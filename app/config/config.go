package config

import (
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/redis/go-redis/v9"
	"github.com/trycourier/courier-go/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	Mc      *mongo.Client
	Store   *session.Store
	Redis   *redis.Client
	Courier *courier.Client
}
