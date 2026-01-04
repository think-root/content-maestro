package store

type CollectSettings struct {
	MaxRepos           int    `json:"max_repos"`
	Since              string `json:"since"`
	SpokenLanguageCode string `json:"spoken_language_code"`
}
