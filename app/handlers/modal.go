package handlers

import (
	"context"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func HandleSubmitResultModal(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		test_id, err := primitive.ObjectIDFromHex(c.Query("test_id"))
		if err != nil {
			return err
		}
		coll := cfg.Mc.Database("primary").Collection("tests")
		var test schema.Test
		if err := coll.FindOne(context.Background(), bson.D{{"_id", test_id}}).Decode(&test); err != nil {
			return err
		}
		return c.Render("modals/submit_result", fiber.Map{
			"Test": test,
		})
	}
}
