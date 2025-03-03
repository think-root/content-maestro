package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type updateResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func UpdateRepositoryPosted(url string, posted bool) (bool, error) {
	apiURL := os.Getenv("CONTENT_ALCHEMIST_URL") + "/think-root/api/update-posted/"

	payload := strings.NewReader(fmt.Sprintf(`{
		"url": "%s",
		"posted": %t
	}`, url, posted))

	req, err := http.NewRequest("PATCH", apiURL, payload)
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("CONTENT_ALCHEMIST_BEARER"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	var response updateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("error decoding response: %v", err)
	}

	return response.Status == "ok", nil
}
