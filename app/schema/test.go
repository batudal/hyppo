package schema

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Test struct {
	ObjectId         primitive.ObjectID `bson:"_id"`
	ModelId          primitive.ObjectID `bson:"modelid"`
	MethodId         primitive.ObjectID `bson:"methodid"`
	UserId           primitive.ObjectID `bson:"userid"`
	Project          string             `bson:"project"`
	Title            string             `bson:"title"`
	StartDate        primitive.DateTime `bson:"startdate"`
	EndDate          primitive.DateTime `bson:"enddate"`
	Status           string             `bson:"status"`
	State            string             `bson:"state"`
	TargetAudience   string             `bson:"targetaudience"`
	ProblemStatement string             `bson:"problemstatement"`
	ProposedSolution string             `bson:"proposedsolution"`
	KPI              string             `bson:"kpi"`
	SuccessCriteria  float64            `bson:"successcriteria"`
	Completed        bool               `bson:"completed"`
	Result           float64            `bson:"result"`
	CreatedAt        primitive.DateTime `bson:"createdat"`
	UpdatedAt        primitive.DateTime `bson:"updatedat"`
}
