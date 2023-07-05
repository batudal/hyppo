package handlers

import (
	"context"
	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func HandleWelcomeModal(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("modals/welcome", fiber.Map{})
	}
}

func HandleMembershipModal(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		coll := cfg.Mc.Database("primary").Collection("memberships")
		var memberships []schema.Membership
		cursor, err := coll.Find(context.TODO(), bson.D{{}})
		if err != nil {
			return err
		}
		if err = cursor.All(context.Background(), &memberships); err != nil {
			return err
		}
		return c.Render("modals/membership", fiber.Map{
			"Memberships": memberships,
		})
	}
}
