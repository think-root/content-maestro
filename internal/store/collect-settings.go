package store

type CollectSettings struct {
	MaxRepos           int    `json:"max_repos"`
	Resource           string `json:"resource"`
	Since              string `json:"since"`
	SpokenLanguageCode string `json:"spoken_language_code"`
	Period             string `json:"period"`
	Language           string `json:"language"`
}
