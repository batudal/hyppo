package schema

import (
	"html/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Model struct {
	ObjectId    primitive.ObjectID `bson:"_id"`
	CategoryId  primitive.ObjectID `bson:"categoryid"`
	Name        string             `bson:"name"`
	Flatname    string             `bson:"flatname"`
	ReviewCount int64              `bson:"reviewcount"`
	TestCount   int64              `bson:"testcount"`
	Companies   []string           `bson:"companies"`
	CreatedAt   primitive.DateTime `bson:"createdat"`
	UpdatedAt   primitive.DateTime `bson:"updatedat"`
	Content     template.HTML      `bson:"content"`
}

type Feed struct {
	Page   int64   `bson:"page"`
	SortBy string  `bson:"sortby"`
	Models []Model `bson:"models"`
}

func (m Model) IsLast(i int) bool {
	return i == 3
}

func (m Model) Increment(i int64) int64 {
	return i + 1
}
// func (m Model) ParseDescription() template.HTML {
// 	buf := mdToHTML([]byte(m.Description))
// 	return template.HTML(buf)
// }

func mdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return markdown.Render(doc, renderer)
}
