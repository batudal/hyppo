package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/batudal/hyppo/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v74/webhook"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleSubscribe(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		event, err := webhook.ConstructEvent(c.Body(), c.GetReqHeaders()["Stripe-Signature"], os.Getenv("STRIPE_SUBSCRIBE_SECRET"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Error verifying webhook signature")
		}
		switch event.Type {
		case "invoice.paid":
			if event.Data.Object["billing_reason"] == "subscription_create" {
				coll := cfg.Mc.Database("primary").Collection("users")
				var user schema.User
				err := coll.FindOne(context.Background(), bson.D{{"email", event.Data.Object["customer_email"].(string)}}).Decode(&user)
				if err == mongo.ErrNoDocuments {
					user := schema.User{
						ObjectId:     primitive.NewObjectID(),
						Email:        event.Data.Object["customer_email"].(string),
						Name:         event.Data.Object["customer_name"].(string),
						Membership:   true,
						MembershipAt: time.Now().Unix(),
						StripeId:     event.Data.Object["customer"].(string),
					}
					_, err = coll.InsertOne(context.Background(), user)
					if err != nil {
						return err
					}
					return c.SendStatus(fiber.StatusOK)
				} else if err != nil {
					return err
				}
				filter := bson.D{{"email", event.Data.Object["customer_email"].(string)}}
				update := bson.D{
					{"$set",
						bson.D{
							{"membership", true},
							{"membershipat", time.Now().Unix()},
							{"stripeid", event.Data.Object["customer"].(string)}}}}
				_, err = coll.UpdateOne(context.Background(), filter, update)
			}
		default:
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func HandleSignup(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user schema.User
		coll := cfg.Mc.Database("primary").Collection("users")
		err := coll.FindOne(context.Background(), bson.D{{"email", c.FormValue("email")}}).Decode(&user)
		if err != mongo.ErrNoDocuments {
			return c.Render("partials/account-exists", fiber.Map{
				"Email": c.FormValue("email"),
			})
		}
		user.ObjectId = primitive.NewObjectID()
		user.Name = c.FormValue("name")
		user.Email = c.FormValue("email")
		user.CreatedAt = time.Now().Unix()
		user.UpdatedAt = time.Now().Unix()
		rand.Seed(time.Now().UnixNano())
		magic := utils.RandomString(16)
		unverified_users_coll := cfg.Mc.Database("secondary").Collection("unverified-users")
		_, err = unverified_users_coll.InsertOne(context.Background(), user)
		if err != nil {
			return err
		}
		params := map[string]string{
			"name":  c.FormValue("name"),
			"magic": c.BaseURL() + "/magic" + "/" + user.Email + "/" + magic,
		}
		err = utils.SendEmail(cfg.Courier, user.Email, "DXNPQFBTGTMPPZPXNP325NV73PHN", params)
		if err != nil {
			return err
		}
		return c.Render("partials/otp", fiber.Map{
			"Email": user.Email,
		})
	}
}

func HandleCancelMember(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		event, err := webhook.ConstructEvent(c.Body(), c.GetReqHeaders()["Stripe-Signature"], os.Getenv("STRIPE_CANCEL_SECRET"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Error verifying webhook signature")
		}
		fmt.Println("event type:", event.Type)
		switch event.Type {
		case "subscription_schedule.aborted", "subscription_schedule.canceled":
			fmt.Println("data ojb:", event.Data.Object)
			coll := cfg.Mc.Database("primary").Collection("users")
			filter := bson.D{{"stripeid", event.Data.Object["customer"].(string)}}
			update := bson.D{{"$set", bson.D{{"membership", false}}}}
			_, err = coll.UpdateOne(context.Background(), filter, update)
			if err != nil {
				return err
			}
		default:
		}
		return c.SendStatus(fiber.StatusOK)
	}
}
