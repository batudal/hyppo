package main

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"html/template"
)

func (m BusinessModel) IsLast(i int) bool {
	return i == 3
}

func (m BusinessModel) Increment(i int64) int64 {
	return i + 1
}

func (m BusinessModel) ParseDescription() template.HTML {
	buf := mdToHTML([]byte(m.Description))
	return template.HTML(buf)
}

func mdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return markdown.Render(doc, renderer)
}
