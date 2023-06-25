package handlers

import (
	"context"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/middleware"
	"github.com/batudal/hyppo/schema"
	"github.com/batudal/hyppo/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleMagicLink(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user schema.User
		coll := cfg.Mc.Database("primary").Collection("users")
		err := coll.FindOne(context.Background(), bson.D{{"email", c.FormValue("email")}}).Decode(&user)
		if err != nil {
			return c.Render("partials/no-account", fiber.Map{
				"Email": c.FormValue("email"),
			})
		}
		magic := utils.RandomString(16)
		if err := utils.SaveMagic(cfg.Redis, user.Email, magic); err != nil {
			return err
		}
		params := map[string]string{
			"name":  user.Name,
			"magic": c.BaseURL() + "/login" + "/" + user.Email + "/" + magic,
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

func HandleLogout(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		sess.Destroy()
		return c.Redirect("/")
	}
}

func HandleLogin(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		magic := c.Params("magic")
		email := c.Params("email")
		if err := utils.VerifyMagic(cfg.Redis, email, magic); err != nil {
			return err
		}
		var user schema.User
		coll := cfg.Mc.Database("primary").Collection("users")
		err := coll.FindOne(context.Background(), bson.D{{"email", email}}).Decode(&user)
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		sess.Set("user", user)
		if err := sess.Save(); err != nil {
			return err
		}
		return c.Redirect("/")
	}
}

func HandleGoogleLogin(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		claims, err := middleware.AuthorizeGoogleJWT(c)
		if err != nil {
			return err
		}
		name := claims.FirstName + " " + claims.LastName
		email := claims.Email
		var user schema.User
		coll := cfg.Mc.Database("primary").Collection("users")
		err = coll.FindOne(context.Background(), bson.D{{"email", email}}).Decode(&user)
		if err == mongo.ErrNoDocuments {
			result, err := coll.InsertOne(context.Background(), bson.D{
				{"name", name},
				{"email", email},
			})
			if err != nil {
				return err
			}
			err = coll.FindOne(context.Background(), bson.D{{"_id", result.InsertedID}}).Decode(&user)
			if err != nil {
				return err
			}
		}
		sess.Set("user", user)
		if err := sess.Save(); err != nil {
			return err
		}
		return c.Redirect("/")
	}
}