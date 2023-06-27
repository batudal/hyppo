package main

import (
	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/handlers"
	"github.com/batudal/hyppo/middleware"
	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App, cfg config.Config) {
	app.Get("/", handlers.IndexPage(cfg))
	app.Get("/welcome", handlers.HandleWelcomePage(cfg)) // stripe callback
	app.Post("/signup", handlers.HandleSignup(cfg))
	app.Post("/magiclink", handlers.HandleMagicLink(cfg))
	app.Get("/login/:email/:magic", handlers.HandleLogin(cfg))
	app.Post("/login/google", handlers.HandleGoogleLogin(cfg))
	app.Post("/hooks/subscribe", handlers.HandleSubscribe(cfg)) // stripe webhook
	app.Post("/hooks/cancel", handlers.HandleCancelMember(cfg)) // stripe webhook
	app.Get("/models", handlers.HandleGetModels(cfg))
	app.Get("/business-model/:flatname", handlers.ModelPage(cfg))
	app.Get("/partials/model", handlers.HandleGetModel(cfg))
	app.Get("/review/new", middleware.Authorize(cfg), handlers.NewReview(cfg))
	app.Get("/logout", middleware.Authorize(cfg), handlers.HandleLogout(cfg))
	app.Get("/modals/welcome", handlers.HandleWelcomeModal(cfg))
	app.Get("/modals/reviews", middleware.Authorize(cfg), handlers.HandleReviewsModal(cfg))
	app.Post("/search", handlers.HandleSearch(cfg))
	app.Post("/search/reviews", handlers.HandleSearchReviews(cfg))
	app.Post("/review", middleware.Authorize(cfg), handlers.HandleNewReview(cfg))
	app.Patch("/newsletter/model/subscribe", handlers.HandleSubscribeModels(cfg))
	app.Patch("/newsletter/model/cancel", handlers.HandleUnsubscribeModels(cfg))
	app.Static("/assets", "./assets")
	app.Static("/", "./public")
	app.Listen(":80")
}
