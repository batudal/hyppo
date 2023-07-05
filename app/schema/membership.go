package schema

type Membership struct {
	Type       string   `json:"type"`
	Price      int64    `json:"price"`
	Period     string   `json:"period"`
	Features   []string `json:"features"`
	Url        string   `json:"url"`
	ButtonType string   `json:"buttonType"`
	ButtonText string   `json:"buttonText"`
}
