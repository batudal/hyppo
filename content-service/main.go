package main

import (
	"context"
	"os"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Model struct {
	Name     string
	Flatname string
	Content  string
}

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		panic(err)
	}
	mng, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	c := colly.NewCollector(
		colly.MaxDepth(1),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})
	var visited_urls []string
	c.OnHTML(".list-unstyled", func(e *colly.HTMLElement) {
		links := e.ChildAttrs("a", "href")
		for _, link := range links {
			if startsWith(link, "/plays/business-model/aikido") && !contains(visited_urls, link) {
				visited_urls = append(visited_urls, link)
				println("Visiting", e.Request.AbsoluteURL(link))
				c.Visit(e.Request.AbsoluteURL(link))
			}
		}
	})
	elements := []string{}
	c.OnHTML("main", func(e *colly.HTMLElement) {
		if e.Request.URL.Path != "/playbooks/" {
			h1 := e.ChildText("h1")
			if h1 != "" {
				elements = append(elements, h1)
			}
			e.ForEach("p", func(_ int, el *colly.HTMLElement) {
				if el.Text != "" &&
					!el.DOM.Parent().HasClass("card-header") &&
					!el.DOM.HasClass("lead") &&
					!el.DOM.HasClass("mt-2") {
					elements = append(elements, el.Text)
				}
			})
		}
	})
	c.Visit("https://learningloop.io/playbooks/")
	prompt := "\nWrite 3 cons and 3 pros of this business model in html using only h2 ol and li tags. Parent elements are not needed. After this write a 2 paragraph summary of the business model using only p tags."
	content, err := AskAI(elements, prompt)
	if err != nil {
		panic(err)
	}
	Write(mng, Model{
		Name:     "🥊 Aikido",
		Flatname: "aikido",
		Content:  content,
	})
}

func Write(mng *mongo.Client, model Model) {
	coll := mng.Database("primary").Collection("business-models")
	_, err := coll.InsertOne(context.Background(), model)
	if err != nil {
		panic(err)
	}
}

func AskAI(elements []string, prompt string) (string, error) {
	consolidated := consolidate(append(elements, prompt))
	client := openai.NewClient(os.Getenv("OPENAI_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: consolidated,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	println(resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}

func consolidate(elements []string) string {
	var consolidated string
	for _, element := range elements {
		consolidated += element + "\n"
	}
	return consolidated
}

func startsWith(text string, pre string) bool {
	return len(text) >= len(pre) && text[0:len(pre)] == pre
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
