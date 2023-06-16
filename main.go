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
		Views: engine,
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("pages/index", fiber.Map{
			"Title": "Hello, Index!",
			"Tags": []string{
				"🔬 Most searched",
				"🧪 Most tested",
				"⚡️ Top rated",
				"🥰️ Most popular",
				"👀 Most recent",
				"🎙 Most talked about",
				"👩‍💻️ Most used",
				"✨ Most rated"},
		}, "layouts/main")
	})
	app.Get("/get-test", func(c *fiber.Ctx) error {
		time.Sleep(1 * time.Second)
		return c.Render("partials/test", fiber.Map{
			"Time": time.Now().Format("2006-01-02 15:04:05"),
		})
	})
	app.Static("/assets", "./assets")
	app.Listen(":80")
}
