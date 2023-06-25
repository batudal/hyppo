package handlers

import (
	"github.com/batudal/hyppo/config"
	"github.com/gofiber/fiber/v2"
)

func HandleWelcomeModal(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("modals/welcome", fiber.Map{})
	}
}
