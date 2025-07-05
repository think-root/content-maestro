package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func init() {
	getRepositoryUrl = os.Getenv("CONTENT_ALCHEMIST_URL") + "/think-root/api/get-repository/"
	bearerToken = "Bearer " + os.Getenv("CONTENT_ALCHEMIST_BEARER")
}

type repo struct {
	ID         int     `json:"id"`
	Posted     bool    `json:"posted"`
	URL        string  `json:"url"`
	Text       string  `json:"text"`
	DateAdded  string  `json:"date_added"`
	DatePosted *string `json:"date_posted"`
}

type repositoryResponse struct {
	Data struct {
		All        int    `json:"all"`
		Posted     int    `json:"posted"`
		Unposted   int    `json:"unposted"`
		Items      []repo `json:"items"`
		Page       int    `json:"page"`
		PageSize   int    `json:"page_size"`
		TotalPages int    `json:"total_pages"`
		TotalItems int    `json:"total_items"`
	} `json:"data"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func GetRepository(limit int, posted bool, sort_order, sort_by string, textLanguage ...string) (*repositoryResponse, error) {
	var lang string
	if len(textLanguage) > 0 && textLanguage[0] != "" {
		lang = textLanguage[0]
	}

	response, err := makeRepositoryRequest(limit, posted, sort_order, sort_by, lang)
	if err != nil {
		return nil, err
	}

	if response.Status == "error" && strings.Contains(response.Message, "no text available for language") && lang != "uk" {
		return makeRepositoryRequest(limit, posted, sort_order, sort_by, "uk")
	}

	return response, nil
}

func makeRepositoryRequest(limit int, posted bool, sort_order, sort_by, textLanguage string) (*repositoryResponse, error) {
	payloadStr := fmt.Sprintf(`{
			"limit": %d,
			"posted": %t,
			"sort_order": "%s",
			"sort_by": "%s",
			"text_language": "%s"
		}`, limit, posted, sort_order, sort_by, textLanguage)

	payload := strings.NewReader(payloadStr)

	req, err := http.NewRequest(http.MethodPost, getRepositoryUrl, payload)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header = http.Header{
		"Accept":        {"*/*"},
		"Connection":    {"keep-alive"},
		"Content-Type":  {"application/json"},
		"Authorization": {bearerToken},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	var response repositoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &response, nil
}
