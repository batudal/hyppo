package utils

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/trycourier/courier-go/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func CheckMembership(db *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user schema.User
		coll := db.Database("primary").Collection("users")
		err := coll.FindOne(context.Background(), c.Locals("user")).Decode(&user)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "You are not a member of this organization")
		}
		return c.Next()
	}
}

func RandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func SendEmail(client *courier.Client, to string, template string, data map[string]string) error {
	_, err := client.SendMessage(
		context.Background(),
		courier.SendMessageRequestBody{
			Message: map[string]interface{}{
				"to": map[string]string{
					"email": to,
				},
				"brand_id": "ZE1HM57RN74BBGPKNYQJ797YYZPN",
				"template": template,
				"data":     data,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func SaveMagic(client *redis.Client, mail_address string, magic string) error {
	key := "magic:" + mail_address
	err := client.Set(context.Background(), key, magic, 300*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func VerifyMagic(client *redis.Client, mail_address string, magic string) error {
	key := "magic:" + mail_address
	magic_db, err := client.Get(context.Background(), key).Result()
	if err != nil {
		return err
	}
	if magic_db != magic {
		return errors.New("")
	}
	return nil
}
