package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v74/webhook"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Authorize(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.store.Get(c)
		if err != nil {
			return err
		}
		if sess.Get("user_id") == nil {
			return c.Render("modals/welcome", fiber.Map{})
		}
		user_id, err := primitive.ObjectIDFromHex(sess.Get("user_id").(string))
		if err != nil {
			return err
		}
		var user User
		coll := cfg.mc.Database("primary").Collection("users")
		filter := bson.D{{"_id", user_id}}
		err = coll.FindOne(context.Background(), filter).Decode(&user)
		if err != nil {
			return err
		}
		c.Locals("user", user)
		return c.Next()
	}
}

func NewReview(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(User)
		if !user.Membership {
			return fiber.NewError(fiber.StatusForbidden, "You must be a member to review models.")
		}
		model_id, err := primitive.ObjectIDFromHex(c.Query("model_id"))
		if err != nil {
			return err
		}
		fmt.Println("model_id: ", model_id)
		var model BusinessModel
		models_coll := cfg.mc.Database("primary").Collection("business-models")
		filter := bson.D{{"_id", model_id}}
		err = models_coll.FindOne(context.Background(), filter).Decode(&model)
		if err != nil {
			return err
		}
		return c.Render("partials/new_review", fiber.Map{
			"Model": model,
			"User":  user,
		})
	}
}

// func NextReviews(cfg Config) fiber.Handler {
//   return func(c *fiber.Ctx) error {
//     createdat, err := strconv.Atoi(c.Query("createdat"))
//     if err != nil {
//       return err
//     }

//   }
// }

func HandleReviewsModal(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type AuthoredReview struct {
			Review Review
			Author User
		}
		coll := cfg.mc.Database("primary").Collection("reviews")
		model_id, err := primitive.ObjectIDFromHex(c.Query("model_id"))
		if err != nil {
			return err
		}
		filter := bson.D{{"modelid", model_id}}
		opts := options.Find().SetSort(bson.D{{"createdat", -1}}).SetLimit(4).SetSkip(0)
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err != nil {
			return err
		}
		var reviews []Review
		if err = cursor.All(context.Background(), &reviews); err != nil {
			return err
		}
		wg := sync.WaitGroup{}
		authored_reviews := make([]AuthoredReview, len(reviews))
		for i, review := range reviews {
			wg.Add(1)
			go func(i int, review Review, authored_reviews []AuthoredReview, wg *sync.WaitGroup) {
				authored_reviews[i].Review = review
				fmt.Println(i, " ", review)
				filter = bson.D{{"_id", review.UserId}}
				err = cfg.mc.Database("primary").Collection("users").FindOne(context.Background(), filter).Decode(&authored_reviews[i].Author)
				if err != nil {
					authored_reviews[i].Author = User{
						Name: "Deleted User",
					}
				}
				wg.Done()
			}(i, review, authored_reviews, &wg)
		}
		wg.Wait()
		coll = cfg.mc.Database("primary").Collection("business-models")
		filter = bson.D{{"_id", model_id}}
		var model BusinessModel
		err = coll.FindOne(context.Background(), filter).Decode(&model)
		if err != nil {
			return err
		}
		return c.Render("modals/reviews", fiber.Map{
			"Model":   model,
			"Reviews": authored_reviews,
		})
	}
}

func HandleComment(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.FormValue("comment") == "" {
			return c.Render("partials/comment", fiber.Map{
				"Error": "Comment cannot be empty",
			})
		}
		fmt.Println("model_id: ", c.FormValue("model_id"))
		model_id, err := primitive.ObjectIDFromHex(c.FormValue("model_id"))
		if err != nil {
			return err
		}
		user_id := c.Locals("user_id").(primitive.ObjectID)
		var user User
		filter := bson.D{{"_id", user_id}}
		err = cfg.mc.Database("primary").Collection("users").FindOne(context.Background(), filter).Decode(&user)
		if err != nil {
			return err
		}
		review := Review{
			ObjectId:  primitive.NewObjectID(),
			ModelId:   model_id,
			UserId:    user_id,
			Comment:   c.FormValue("comment"),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		reviews_coll := cfg.mc.Database("primary").Collection("reviews")
		_, err = reviews_coll.InsertOne(context.Background(), review)
		if err != nil {
			return err
		}
		models_coll := cfg.mc.Database("primary").Collection("business-models")
		filter = bson.D{{"_id", model_id}}
		update := bson.D{{"$set", bson.D{{"latestreview", review.Comment}}}}
		_, err = models_coll.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}
		update_review_count := bson.D{{"$inc", bson.D{{"reviewcount", 1}}}}
		_, err = models_coll.UpdateOne(context.Background(), filter, update_review_count)
		if err != nil {
			return err
		}
		return c.Render("partials/comment", fiber.Map{
			"User":   user,
			"Review": review,
		})
	}
}

func HandleSearch(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.FormValue("query") == "" {
			feed := Feed{
				Page:   1,
				SortBy: "createdat",
			}
			filter := bson.D{}
			opts := options.
				Find().
				SetSort(bson.D{{"createdat", -1}}).
				SetLimit(4)
			coll := cfg.mc.Database("primary").Collection("business-models")
			cursor, err := coll.Find(context.Background(), filter, opts)
			if err = cursor.All(context.TODO(), &feed.Models); err != nil {
				panic(err)
			}
			return c.Render("partials/business-model", fiber.Map{
				"Feed": feed,
			})
		}
		coll := cfg.mc.Database("primary").Collection("business-models")
		filter := bson.D{{"$text", bson.D{{"$search", c.FormValue("query")}}}}
		var models []BusinessModel
		cursor, err := coll.Find(context.TODO(), filter)
		if err != nil {
			return err
		}
		if err = cursor.All(context.Background(), &models); err != nil {
			return err
		}
		return c.Render("partials/business-models", fiber.Map{
			"Models": models,
		})
	}
}

func HandleLogin(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user User
		coll := cfg.mc.Database("primary").Collection("users")
		err := coll.FindOne(context.Background(), bson.D{{"email", c.FormValue("email")}}).Decode(&user)
		if err != nil {
			return c.Render("partials/no-account", fiber.Map{
				"Email": c.FormValue("email"),
			})
		}
		otp, err := SetOTP(cfg.redis, user.Email)
		if err != nil {
			return err
		}
		params := map[string]string{
			"name":  user.Name,
			"magic": c.BaseURL() + "/magic" + "/" + user.Email + "/" + otp,
		}
		err = sendEmail(cfg.courier, user.Email, "DXNPQFBTGTMPPZPXNP325NV73PHN", params)
		if err != nil {
			return err
		}
		return c.Render("partials/otp", fiber.Map{
			"Email": user.Email,
		})
	}
}

func HandleLogout(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.store.Get(c)
		if err != nil {
			return err
		}
		sess.Destroy()
		return c.Redirect("/")
	}
}

func HandleMagic(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		magic := c.Params("magic")
		email := c.Params("email")
		if err := VerifyOTP(cfg.redis, email, magic); err != nil {
			return err
		}
		var user User
		coll := cfg.mc.Database("primary").Collection("users")
		err := coll.FindOne(context.Background(), bson.D{{"email", email}}).Decode(&user)
		sess, err := cfg.store.Get(c)
		if err != nil {
			return err
		}
		sess.Set("user_id", user.ObjectId.Hex())
		if err := sess.Save(); err != nil {
			return err
		}
		return c.Redirect("/")
	}
}

func HandleGetModels(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			return err
		}
		order, err := strconv.Atoi(c.Query("order"))
		if err != nil {
			return err
		}
		feed := Feed{
			Page:   int64(page),
			SortBy: c.Query("sortby"),
		}
		filter := bson.D{}
		opts := options.
			Find().
			SetSort(bson.D{{c.Query("sortby"), order}}).
			SetLimit(4).
			SetSkip(int64(page-1) * 4)
		coll := cfg.mc.Database("primary").Collection("business-models")
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err = cursor.All(context.TODO(), &feed.Models); err != nil {
			panic(err)
		}
		return c.Render("partials/business-model", fiber.Map{
			"Feed": feed,
		})
	}
}

func IndexPage(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.store.Get(c)
		if err != nil {
			panic(err)
		}
		var user User
		session_email := sess.Get("user_email")
		if session_email != nil {
			coll := cfg.mc.Database("primary").Collection("users")
			err := coll.FindOne(context.Background(), bson.D{{"email", session_email}}).Decode(&user)
			if err != nil {
				return err
			}
		}
		feed := Feed{
			Page:   1,
			SortBy: "createdat",
		}
		filter := bson.D{}
		opts := options.Find().SetSort(bson.D{{"createdat", -1}}).SetLimit(4)
		coll := cfg.mc.Database("primary").Collection("business-models")
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err = cursor.All(context.TODO(), &feed.Models); err != nil {
			return err
		}
		return c.Render("pages/index", fiber.Map{
			"User": user,
			"Feed": feed,
		}, "layouts/main")
	}
}

func HandleCreateMember(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		event, err := webhook.ConstructEvent(c.Body(), c.GetReqHeaders()["Stripe-Signature"], os.Getenv("STRIPE_CREATEUSER_SECRET"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Error verifying webhook signature")
		}
		switch event.Type {
		case "invoice.paid":
			if event.Data.Object["billing_reason"] == "subscription_create" {
				coll := cfg.mc.Database("primary").Collection("users")
				var user User
				err := coll.FindOne(context.Background(), bson.D{{"email", event.Data.Object["customer_email"].(string)}}).Decode(&user)
				if err == mongo.ErrNoDocuments {
					user := User{
						ObjectId:     primitive.NewObjectID(),
						Email:        event.Data.Object["customer_email"].(string),
						Name:         event.Data.Object["customer_name"].(string),
						Membership:   true,
						MembershipAt: time.Now().Unix(),
						StripeId:     event.Data.Object["customer"].(string),
					}
					_, err = coll.InsertOne(context.Background(), user)
					if err != nil {
						return err
					}
					return c.SendStatus(fiber.StatusOK)
				} else if err != nil {
					return err
				}
				filter := bson.D{{"email", event.Data.Object["customer_email"].(string)}}
				update := bson.D{
					{"$set",
						bson.D{
							{"membership", true},
							{"membershipat", time.Now().Unix()},
							{"stripeid", event.Data.Object["customer"].(string)}}}}
				_, err = coll.UpdateOne(context.Background(), filter, update)
			}
		default:
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func HandleCancelMember(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		event, err := webhook.ConstructEvent(c.Body(), c.GetReqHeaders()["Stripe-Signature"], os.Getenv("STRIPE_CREATEUSER_SECRET"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Error verifying webhook signature")
		}
		fmt.Println("event type:", event.Type)
		switch event.Type {
		case "subscription_schedule.aborted", "subscription_schedule.canceled":
			fmt.Println("data ojb:", event.Data.Object)
			coll := cfg.mc.Database("primary").Collection("users")
			filter := bson.D{{"stripeid", event.Data.Object["customer"].(string)}}
			update := bson.D{{"$set", bson.D{{"membership", false}}}}
			_, err = coll.UpdateOne(context.Background(), filter, update)
			if err != nil {
				return err
			}
		default:
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func HandleWelcomePage(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("pages/index", fiber.Map{
			"Modal": "welcome",
		}, "layouts/main")
	}
}

func HandleWelcomeModal(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("modals/welcome", fiber.Map{})
	}
}
