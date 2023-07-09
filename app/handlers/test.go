package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/batudal/hyppo/config"
	"github.com/batudal/hyppo/schema"
	"github.com/batudal/hyppo/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleDeleteTest(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		testid, err := primitive.ObjectIDFromHex(c.Query("test_id"))
		if err != nil {
			return err
		}
		coll := cfg.Mc.Database("primary").Collection("tests")
		filter := bson.D{{"_id", testid}}
		_, err = coll.DeleteOne(context.Background(), filter)
		if err != nil {
			return err
		}
		c.Append("HX-Redirect", "/tests")
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func HandleCompleteTest(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		testid, err := primitive.ObjectIDFromHex(c.Query("test_id"))
		if err != nil {
			return err
		}
		result := c.FormValue("result")
		coll := cfg.Mc.Database("primary").Collection("tests")
		filter := bson.D{{"userid", user.ObjectId}, {"_id", testid}}
		update := bson.D{{"$set", bson.D{
			{"state", "completed"},
			{"result", result},
			{"updatedat", primitive.NewDateTimeFromTime(time.Now())},
		}}}
		_, err = coll.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}
		return c.Redirect("/tests/completed")
	}
}

func HandleNewTest(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		new_test := schema.Test{
			ObjectId:         primitive.NewObjectID(),
			UserId:           user.ObjectId,
			Title:            "New Test",
			StartDate:        primitive.NewDateTimeFromTime(time.Now()),
			EndDate:          primitive.NewDateTimeFromTime(time.Now()),
			Project:          "",
			TargetAudience:   "",
			ProblemStatement: "",
			ProposedSolution: "",
			KPI:              "",
			SuccessCriteria:  0,
			Status:           "public",
			State:            "ongoing",
			CreatedAt:        primitive.NewDateTimeFromTime(time.Now()),
			UpdatedAt:        primitive.NewDateTimeFromTime(time.Now()),
		}
		coll := cfg.Mc.Database("primary").Collection("tests")
		_, err = coll.InsertOne(context.Background(), new_test)
		if err != nil {
			return err
		}
		models, err := utils.GetAllModels(cfg)
		if err != nil {
			return err
		}
		methods, err := utils.GetAllMethods(cfg)
		if err != nil {
			return err
		}
		return c.Render("pages/tests/edit", fiber.Map{
			"User":    user,
			"Models":  models,
			"Methods": methods,
			"Test":    new_test,
		})
	}
}

func HandleEditTest(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := cfg.Store.Get(c)
		if err != nil {
			return err
		}
		user := sess.Get("user").(*schema.User)
		testid, err := primitive.ObjectIDFromHex(c.Query("test_id"))
		if err != nil {
			return err
		}
		coll := cfg.Mc.Database("primary").Collection("tests")
		filter := bson.D{{"userid", user.ObjectId}, {"_id", testid}}
		start := c.FormValue("start")
		start_int, err := strconv.Atoi(start)
		if err != nil {
			return err
		}
		startdate := time.Unix(int64(start_int)/1000, 0)
		startdateobj := primitive.NewDateTimeFromTime(startdate)
		end := c.FormValue("end")
		end_int, err := strconv.Atoi(end)
		if err != nil {
			return err
		}
		enddate := time.Unix(int64(end_int)/1000, 0)
		enddateobj := primitive.NewDateTimeFromTime(enddate)
		successcriteria := c.FormValue("successcriteria")
		successcriteria_float, err := strconv.ParseFloat(successcriteria, 64)
		if err != nil {
			return err
		}
		modelname := c.FormValue("model")
		coll_models := cfg.Mc.Database("primary").Collection("business-models")
		filter_models := bson.D{{"name", modelname}}
		var model schema.Model
		err = coll_models.FindOne(context.Background(), filter_models).Decode(&model)
		if err != nil {
			return err
		}
		methodname := c.FormValue("method")
		coll_methods := cfg.Mc.Database("primary").Collection("methods")
		filter_methods := bson.D{{"name", methodname}}
		var method schema.Method
		err = coll_methods.FindOne(context.Background(), filter_methods).Decode(&method)
		if err != nil {
			return err
		}
		test := schema.Test{
			UserId:           user.ObjectId,
			Title:            c.FormValue("title"),
			StartDate:        startdateobj,
			EndDate:          enddateobj,
			Project:          c.FormValue("project"),
			TargetAudience:   c.FormValue("targetaudience"),
			ProblemStatement: c.FormValue("problemstatement"),
			ProposedSolution: c.FormValue("proposedsolution"),
			KPI:              c.FormValue("kpi"),
			SuccessCriteria:  successcriteria_float,
			ModelId:          model.ObjectId,
			MethodId:         method.ObjectId,
			Status:           c.FormValue("status"),
			UpdatedAt:        primitive.NewDateTimeFromTime(time.Now()),
		}
		validator := validator.New()
		err = validator.Struct(test)
		if err != nil {
			return err
		}
		update := bson.D{
			{"$set", bson.D{
				{"userid", user.ObjectId},
				{"title", c.FormValue("title")},
				{"startdate", startdateobj},
				{"enddate", enddateobj},
				{"project", c.FormValue("project")},
				{"targetaudience", c.FormValue("targetaudience")},
				{"problemstatement", c.FormValue("problemstatement")},
				{"proposedsolution", c.FormValue("proposedsolution")},
				{"kpi", c.FormValue("kpi")},
				{"successcriteria", successcriteria_float},
				{"status", c.FormValue("status")},
				{"methodid", method.ObjectId},
				{"modelid", model.ObjectId},
				{"updatedat", primitive.NewDateTimeFromTime(time.Now())},
			}},
		}
		result := coll.FindOneAndUpdate(context.Background(), filter, update)
		if err != nil {
			return err
		}
		var old_test schema.Test
		err = result.Decode(&old_test)
		if err != nil {
			return err
		}
		var updated_test schema.Test
		err = coll.FindOne(context.Background(), filter).Decode(&updated_test)
		if err != nil {
			return err
		}
		utils.UpdateTestCounts(cfg, old_test.ModelId, updated_test.ModelId)
		models, err := utils.GetAllModels(cfg)
		if err != nil {
			return err
		}
		methods, err := utils.GetAllMethods(cfg)
		if err != nil {
			return err
		}
		c.Append("HX-Trigger", "saved")
		return c.Render("pages/tests/edit", fiber.Map{
			"Models":  models,
			"Methods": methods,
			"Test":    updated_test,
		})
	}
}
