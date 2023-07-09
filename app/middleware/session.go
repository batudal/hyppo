package middleware

import (
	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/gofiber/fiber/v2"
)

func Authorize(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		if sess.Get("user") == nil {
			c.Append("HX-Retarget", "body")
			c.Append("HX-Reswap", "beforeend")
			return c.Render("modals/welcome", fiber.Map{})
		}
		c.Locals("user", sess.Get("user").(*schema.User))
		return c.Next()
	}
}

func AuthorizeMember(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*schema.User)
		if !user.Membership {
			c.Append("HX-Retarget", "body")
			c.Append("HX-Reswap", "beforeend")
			return c.Render("modals/membership", fiber.Map{})
		}
		return c.Next()
	}
}
