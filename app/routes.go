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
	app.Get("/partials/review/edit", handlers.EditReview(cfg))
	app.Get("/review/new", middleware.Authorize(cfg), handlers.NewReview(cfg))
	app.Post("/review/helpful", middleware.Authorize(cfg), handlers.HandleHelpfulReview(cfg))
	app.Post("/review/unhelpful", middleware.Authorize(cfg), handlers.HandleUnhelpfulReview(cfg))
	app.Get("/logout", middleware.Authorize(cfg), handlers.HandleLogout(cfg))
	app.Get("/modals/welcome", handlers.HandleWelcomeModal(cfg))
	app.Get("/modals/membership", handlers.HandleMembershipModal(cfg))
	app.Post("/search", handlers.HandleSearch(cfg))
	app.Post("/search/reviews", handlers.HandleSearchReviews(cfg))
	app.Get("/review/discard", handlers.HandleDiscardReview(cfg))
	app.Post("/review", middleware.Authorize(cfg), handlers.HandleNewReview(cfg))
	app.Patch("/review", middleware.Authorize(cfg), handlers.HandleEditReview(cfg))
	app.Delete("/review", middleware.Authorize(cfg), handlers.HandleDeleteReview(cfg))
	app.Patch("/newsletter/model/subscribe", handlers.HandleSubscribeModels(cfg))
	app.Patch("/newsletter/model/cancel", handlers.HandleUnsubscribeModels(cfg))
	app.Get("/model_tabs/model", handlers.ModelTab(cfg))
	app.Get("/model_tabs/reviews", handlers.ReviewsTab(cfg))
	app.Static("/assets", "./assets")
	app.Static("/", "./public")
	app.Listen(":80")
}
