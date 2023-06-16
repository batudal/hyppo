package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	engine := html.New("./views", ".html")
	// Disable this in production
	engine.Reload(true)
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("pages/index", fiber.Map{
			"Title": "Hello, Index!",
		})
	})
	app.Get("/get-test", func(c *fiber.Ctx) error {
		return c.Render("partials/test", fiber.Map{
			"Time": time.Now().Format("2006-01-02 15:04:05"),
		})
	})
	app.Listen(":80")
}
