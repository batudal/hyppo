package schema

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Test struct {
	ObjectId         primitive.ObjectID `bson:"_id"`
	Project          primitive.ObjectID `bson:"projectid"`
	Model            primitive.ObjectID `bson:"modelid"`
	Method           primitive.ObjectID `bson:"methodid"`
	UserId           primitive.ObjectID `bson:"userid"`
	Title            string             `bson:"title"`
	StartDate        int64              `bson:"startdate"`
	EndDate          int64              `bson:"enddate"`
	Private          bool               `bson:"private"`
	TargetAudience   string             `bson:"targetaudience"`
	ProblemStatement string             `bson:"problemstatement"`
	ProposedSolution string             `bson:"proposedsolution"`
	KPI              string             `bson:"kpi"`
	SuccessCriteria  string             `bson:"successcriteria"`
	CreatedAt        int64              `bson:"createdat"`
	UpdatedAt        int64              `bson:"updatedat"`
}
