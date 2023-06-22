package main

import (
	"context"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v74/webhook"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func HandleWelcomeImage(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("partials/welcome-image", fiber.Map{
			"ImageURL": "https://hyppo-files.fra1.cdn.digitaloceanspaces.com/hippo_photo.webp",
		})
	}
}

func HandleLogin(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user User
		coll := cfg.mc.Database("primary").Collection("users")
		err := coll.FindOne(context.Background(), bson.D{{"email", c.FormValue("email")}}).Decode(&user)
		if err != nil {
			return c.Render("partials/no-account", fiber.Map{
				"Email": c.FormValue("email"),
			})
		}
		otp, err := SetOTP(cfg.redis, user.Email)
		if err != nil {
			return err
		}
		params := map[string]string{
			"name":  user.Name,
			"magic": c.BaseURL() + "/magic" + "/" + user.Email + "/" + otp,
		}
		err = sendEmail(cfg.courier, user.Email, "DXNPQFBTGTMPPZPXNP325NV73PHN", params)
		if err != nil {
			return err
		}
		return c.Render("partials/otp", fiber.Map{
			"Email": user.Email,
		})
	}
}

func HandleLogout(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.store.Get(c)
		if err != nil {
			return err
		}
		sess.Destroy()
		return c.Redirect("/")
	}
}

func HandleMagic(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		magic := c.Params("magic")
		email := c.Params("email")
		if err := VerifyOTP(cfg.redis, email, magic); err != nil {
			return err
		}
		var user User
		coll := cfg.mc.Database("primary").Collection("users")
		err := coll.FindOne(context.Background(), bson.D{{"email", email}}).Decode(&user)
		sess, err := cfg.store.Get(c)
		if err != nil {
			return err
		}
		sess.Set("email", user.Email)
		sess.Set("name", user.Name)
		if err := sess.Save(); err != nil {
			return err
		}
		return c.Redirect("/")
	}
}

func HandleGetModels(cfg Config) fiber.Handler {
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
		coll := cfg.mc.Database("primary").Collection("business-models")
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err = cursor.All(context.TODO(), &feed.Models); err != nil {
			panic(err)
		}
		return c.Render("partials/business-model", fiber.Map{
			"Feed": feed,
		})
	}
}

func IndexPage(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.store.Get(c)
		if err != nil {
			panic(err)
		}
		var user User
		session_email := sess.Get("email")
		if session_email != nil {
			coll := cfg.mc.Database("primary").Collection("users")
			err := coll.FindOne(context.Background(), bson.D{{"email", session_email}}).Decode(&user)
			if err != nil {
				return err
			}
		}
		feed := Feed{
			Page:   1,
			SortBy: "createdat",
		}
		filter := bson.D{}
		opts := options.Find().SetSort(bson.D{{"createdat", -1}}).SetLimit(4)
		coll := cfg.mc.Database("primary").Collection("business-models")
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err = cursor.All(context.TODO(), &feed.Models); err != nil {
			return err
		}
		return c.Render("pages/index", fiber.Map{
			"User": user,
			"Feed": feed,
		}, "layouts/main")
	}
}

func HandleCreateUser(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		event, err := webhook.ConstructEvent(c.Body(), c.GetReqHeaders()["Stripe-Signature"], os.Getenv("STRIPE_CREATEUSER_SECRET"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Error verifying webhook signature")
		}
		switch event.Type {
		case "invoice.paid":
			if event.Data.Object["billing_reason"] == "subscription_create" {
				//genrate new mongo id
				id := primitive.NewObjectID()
				user := User{
					ObjectId: id,
					Email:    event.Data.Object["customer_email"].(string),
					Name:     event.Data.Object["customer_name"].(string),
				}
				coll := cfg.mc.Database("primary").Collection("users")
				_, err = coll.InsertOne(context.Background(), user)
				if err != nil {
					return err
				}
			}
		default:
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func HandleWelcomePage(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("pages/index", fiber.Map{
			"Modal": "welcome",
		}, "layouts/main")
	}
}

func HandleWelcomeModal(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("modals/welcome", fiber.Map{})
	}
}
