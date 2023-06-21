package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/mail"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/redis/go-redis/v9"
	"github.com/trycourier/courier-go/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

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

type CustomClaims struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c *CustomClaims) Validate(context.Context) error {
	if c.Email == "" {
		return errors.New("email claim must be present")
	}
	return nil
}

func CheckMembership(db *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user User
		coll := db.Database("primary").Collection("users")
		err := coll.FindOne(context.Background(), c.Locals("user")).Decode(&user)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "You are not a member of this organization")
		}
		return c.Next()
	}
}

func SetOTP(client *redis.Client, mail_address string) (string, error) {
	_, err := mail.ParseAddress(mail_address)
	if err != nil {
		return "", err
	}
	key := "otp:" + mail_address
	fmt.Println("Trying redis")
	_, err = client.Get(context.Background(), key).Result()
	var magic string
	if err == redis.Nil || err == nil {
		rand.Seed(time.Now().UnixNano())
		magic = randomString(16)
		client.Set(context.Background(), key, magic, 30*time.Second)
		fmt.Println("Set redis")
	} else if err != nil {
		return "", err
	}
	return magic, nil
}

func generateSixDigit() int64 {
	value := rand.Intn(900_000) + 100_000
	return int64(value)
}

func randomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func VerifyOTP(client *redis.Client, mail_address string, otp string) error {
	key := "otp:" + mail_address
	otp_db, err := client.Get(context.Background(), key).Result()
	if err != nil {
		return err
	}
	if otp_db != otp {
		return errors.New("")
	}
	return nil
}

func sendEmail(client *courier.Client, to string, template string, data map[string]string) error {
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
