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
		type User struct {
			Name   string
			Avatar string
		}
		return c.Render("pages/index", fiber.Map{
			"User": User{
				Name:   "John",
				Avatar: "https://personal-bucket.fra1.cdn.digitaloceanspaces.com/deadfella.png",
			},
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
