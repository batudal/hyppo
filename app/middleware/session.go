package middleware

import (
	"github.com/batudal/hyppo/config"
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
		return c.Next()
	}
}
