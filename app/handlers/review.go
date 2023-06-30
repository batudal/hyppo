package handlers

import (
	"context"
	"time"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/batudal/hyppo/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
		if !user.Membership {
			return fiber.NewError(fiber.StatusForbidden, "You must be a member to review models.")
		}
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

func HandleReviewsModal(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		if !user.Membership {
			return fiber.NewError(fiber.StatusForbidden, "You must be a member to review models.")
		}
		coll := cfg.Mc.Database("primary").Collection("reviews")
		model_id, err := primitive.ObjectIDFromHex(c.Query("model_id"))
		if err != nil {
			return err
		}
		filter := bson.D{{"modelid", model_id}}
		opts := options.Find().SetSort(bson.D{{"createdat", -1}}).SetSkip(0)
		cursor, err := coll.Find(context.Background(), filter, opts)
		if err != nil {
			return err
		}
		var reviews []schema.Review
		if err = cursor.All(context.Background(), &reviews); err != nil {
			return err
		}
		if len(reviews) == 0 {
			coll = cfg.Mc.Database("primary").Collection("business-models")
			filter = bson.D{{"_id", model_id}}
			var model schema.Model
			err = coll.FindOne(context.Background(), filter).Decode(&model)
			if err != nil {
				return err
			}
			return c.Render("modals/reviews", fiber.Map{
				"Model": model,
			})
		}
		authored_reviews := utils.GetAuthoredReviews(cfg, reviews, *user)
		coll = cfg.Mc.Database("primary").Collection("business-models")
		filter = bson.D{{"_id", model_id}}
		var model schema.Model
		err = coll.FindOne(context.Background(), filter).Decode(&model)
		if err != nil {
			return err
		}
		return c.Render("modals/reviews", fiber.Map{
			"User":    user,
			"Model":   model,
			"Reviews": authored_reviews,
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
		if c.FormValue("comment") == "" {
			return c.Render("partials/comment", fiber.Map{
				"Error": "Comment cannot be empty",
			})
		}
		model_id, err := primitive.ObjectIDFromHex(c.FormValue("model_id"))
		if err != nil {
			return err
		}
		sess, err := cfg.Store.Get(c)
		if err != nil {
			panic(err)
		}
		user := sess.Get("user").(*schema.User)
		review := schema.Review{
			ObjectId:  primitive.NewObjectID(),
			ModelId:   model_id,
			UserId:    user.ObjectId,
			Comment:   c.FormValue("comment"),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		reviews_coll := cfg.Mc.Database("primary").Collection("reviews")
		_, err = reviews_coll.InsertOne(context.Background(), review)
		if err != nil {
			return err
		}
		models_coll := cfg.Mc.Database("primary").Collection("business-models")
		filter := bson.D{{"_id", model_id}}
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
		return c.Render("partials/review/comment", fiber.Map{
			"Author": user,
			"Review": review,
		})
	}
}
