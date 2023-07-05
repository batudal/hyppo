package handlers

import (
	"context"
	"time"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/batudal/hyppo/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ValidateComment(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		comment := c.FormValue("comment")
		var validate = validator.New()
		err := validate.Var(comment, "required,min=20,max=2000")
		if err != nil {
			return c.Render("partials/review/invalid-comment", fiber.Map{
				"Comment": comment,
				"Errors":  err,
			})
		}
		return c.Render("partials/review/validated-comment", fiber.Map{
			"Comment": comment,
		})
	}
}

func HandleDiscardReview(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		coll := cfg.Mc.Database("primary").Collection("reviews")
		review_id, err := primitive.ObjectIDFromHex(c.Query("review_id"))
		if err != nil {
			return err
		}
		var review schema.Review
		filter := bson.D{{"_id", review_id}, {"userid", user.ObjectId}}
		err = coll.FindOne(context.Background(), filter).Decode(&review)
		if err != nil {
			return err
		}
		return c.Render("partials/review/user-comment", fiber.Map{
			"Review": review,
			"Author": user,
		})
	}
}

func HandleEditReview(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		updated_comment := c.FormValue("comment")
		coll := cfg.Mc.Database("primary").Collection("reviews")
		review_id, err := primitive.ObjectIDFromHex(c.Query("review_id"))
		if err != nil {
			return err
		}
		var review schema.Review
		filter := bson.D{{"_id", review_id}, {"userid", user.ObjectId}}
		update := bson.D{{"$set", bson.D{{"comment", updated_comment}}}}
		result := coll.FindOneAndUpdate(context.Background(), filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
		if result.Err() != nil {
			return result.Err()
		}
		result.Decode(&review)
		return c.Render("partials/review/user-comment", fiber.Map{
			"Review": review,
			"Author": user,
		})
	}
}

func EditReview(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		review_id, err := primitive.ObjectIDFromHex(c.Query("review_id"))
		if err != nil {
			return err
		}
		coll := cfg.Mc.Database("primary").Collection("reviews")
		filter := bson.D{{"_id", review_id}, {"userid", user.ObjectId}}
		var review schema.Review
		err = coll.FindOne(context.Background(), filter).Decode(&review)
		if err != nil {
			return err
		}
		return c.Render("partials/review/edit", fiber.Map{
			"User":   user,
			"Review": review,
		})
	}
}

func HandleDeleteReview(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		review_id, err := primitive.ObjectIDFromHex(c.Query("review_id"))
		if err != nil {
			return err
		}
		coll := cfg.Mc.Database("primary").Collection("reviews")
		filter := bson.D{{"_id", review_id}, {"userid", user.ObjectId}}
		var review schema.Review
		err = coll.FindOne(context.Background(), filter).Decode(&review)
		if err != nil {
			return err
		}
		_, err = coll.DeleteOne(context.Background(), filter)
		if err != nil {
			return err
		}
		coll_models := cfg.Mc.Database("primary").Collection("business-models")
		filter = bson.D{{"_id", review.ModelId}}
		update := bson.D{{"$inc", bson.D{{"reviewcount", -1}}}}
		_, err = coll_models.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}
		return c.SendStatus(200)
	}
}

func HandleUnhelpfulReview(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var review schema.Review
		var author schema.User
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		review_id, err := primitive.ObjectIDFromHex(c.Query("review_id"))
		if err != nil {
			return err
		}
		coll := cfg.Mc.Database("primary").Collection("helpfuls")
		filter := bson.D{{"reviewid", review_id}, {"userid", user.ObjectId}}
		_, err = coll.DeleteOne(context.Background(), filter)
		if err != nil {
			return err
		}
		coll_reviews := cfg.Mc.Database("primary").Collection("reviews")
		filter = bson.D{{"_id", review_id}}
		update := bson.D{{"$inc", bson.D{{"helpfulcount", -1}}}}
		result := coll_reviews.FindOneAndUpdate(context.Background(), filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
		if result.Err() != nil {
			return result.Err()
		}
		result.Decode(&review)
		coll_users := cfg.Mc.Database("primary").Collection("users")
		filter = bson.D{{"_id", review.UserId}}
		err = coll_users.FindOne(context.Background(), filter).Decode(&author)
		if err != nil {
			return err
		}
		return c.Render("partials/review/comment", fiber.Map{
			"Review":  review,
			"Author":  author,
			"Helpful": false,
		})
	}
}

func HandleHelpfulReview(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var review schema.Review
		var author schema.User
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		review_id, err := primitive.ObjectIDFromHex(c.Query("review_id"))
		if err != nil {
			return err
		}
		var helpful schema.Helpful
		coll := cfg.Mc.Database("primary").Collection("helpfuls")
		filter := bson.D{{"reviewid", review_id}, {"userid", user.ObjectId}}
		err = coll.FindOne(context.Background(), filter).Decode(&helpful)
		if err == mongo.ErrNoDocuments {
			helpful = schema.Helpful{
				ReviewId: review_id,
				UserId:   user.ObjectId,
			}
			_, err = coll.InsertOne(context.Background(), helpful)
			if err != nil {
				return err
			}
			coll_reviews := cfg.Mc.Database("primary").Collection("reviews")
			filter := bson.D{{"_id", review_id}}
			update := bson.D{{"$inc", bson.D{{"helpfulcount", 1}}}}
			result := coll_reviews.FindOneAndUpdate(context.Background(), filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
			if result.Err() != nil {
				return result.Err()
			}
			result.Decode(&review)
			coll_users := cfg.Mc.Database("primary").Collection("users")
			filter = bson.D{{"_id", review.UserId}}
			err = coll_users.FindOne(context.Background(), filter).Decode(&author)
			if err != nil {
				return err
			}
		}
		return c.Render("partials/review/comment", fiber.Map{
			"Review":  review,
			"Author":  author,
			"Helpful": true,
		})
	}
}

func NewReview(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		model_id, err := primitive.ObjectIDFromHex(c.Query("model_id"))
		if err != nil {
			return err
		}
		var model schema.Model
		models_coll := cfg.Mc.Database("primary").Collection("business-models")
		filter := bson.D{{"_id", model_id}}
		err = models_coll.FindOne(context.Background(), filter).Decode(&model)
		if err != nil {
			return err
		}
		return c.Render("partials/review/new", fiber.Map{
			"Model": model,
			"User":  user,
		})
	}
}

func HandleSearchReviews(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		query := c.FormValue("query")
		coll := cfg.Mc.Database("primary").Collection("reviews")
		model_id, err := primitive.ObjectIDFromHex(c.Query("model_id"))
		if err != nil {
			return err
		}
		var filter bson.D
		if query == "" {
			filter = bson.D{{"modelid", model_id}}
		} else {
			filter = bson.D{{"modelid", model_id}, {"$text", bson.D{{"$search", query}}}}
		}
		opts := options.Find().SetSort(bson.D{{"createdat", -1}})
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err != nil {
			return err
		}
		var reviews []schema.Review
		if err = cursor.All(context.Background(), &reviews); err != nil {
			return err
		}
		if len(reviews) == 0 {
			return c.Render("partials/review/not-found", fiber.Map{
				"Query": query,
			})
		}
		authored_reviews := utils.GetAuthoredReviews(cfg, reviews, *user)
		return c.Render("partials/review/search-reviews", fiber.Map{
			"Reviews": authored_reviews,
		})
	}
}

func HandleNewReview(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		model_id, err := primitive.ObjectIDFromHex(c.FormValue("model_id"))
		if err != nil {
			return err
		}
		var model schema.Model
		models_coll := cfg.Mc.Database("primary").Collection("business-models")
		filter := bson.D{{"_id", model_id}}
		err = models_coll.FindOne(context.Background(), filter).Decode(&model)
		if err != nil {
			return err
		}
		sess, err := cfg.Store.Get(c)
		if err != nil {
			panic(err)
		}
		user := sess.Get("user").(*schema.User)
		if !user.Membership {
			c.Append("HX-Retarget", "body")
			c.Append("HX-Reswap", "beforeend")
			memberships_coll := cfg.Mc.Database("primary").Collection("memberships")
			var memberships []schema.Membership
			cursor, err := memberships_coll.Find(context.TODO(), bson.D{{}})
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
		reviews_coll := cfg.Mc.Database("primary").Collection("reviews")
		result := reviews_coll.FindOne(context.Background(), bson.D{{"modelid", model_id}, {"userid", user.ObjectId}})
		if result.Err() != mongo.ErrNoDocuments {
			c.Append("HX-Retarget", ".review-container")
			c.Append("HX-Reswap", "outerHTML")
			return c.Render("partials/review/new", fiber.Map{
				"Model": model,
				"User":  user,
				"Error": "😓 You have already reviewed this model.",
			})
		}
		var validate = validator.New()
		err = validate.Var(c.FormValue("comment"), "required,min=20,max=2000")
		if err != nil {
			c.Append("HX-Retarget", ".review-container")
			c.Append("HX-Reswap", "outerHTML")
			return c.Render("partials/review/new", fiber.Map{
				"Model": model,
				"User":  user,
				"Error": "😅 Comment must be between 20 and 2000 characters",
			})
		}
		review := schema.Review{
			ObjectId:  primitive.NewObjectID(),
			ModelId:   model_id,
			UserId:    user.ObjectId,
			Comment:   c.FormValue("comment"),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		_, err = reviews_coll.InsertOne(context.Background(), review)
		if err != nil {
			return err
		}
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
		return c.Render("partials/review/user-comment", fiber.Map{
			"Author": user,
			"Review": review,
		})
	}
}
