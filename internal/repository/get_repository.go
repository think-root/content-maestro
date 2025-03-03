package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type repo struct {
	ID     int    `json:"id"`
	Posted bool   `json:"posted"`
	URL    string `json:"url"`
	Text   string `json:"text"`
}

type repositoryResponse struct {
	Data struct {
		All      int    `json:"all"`
		Posted   int    `json:"posted"`
		Unposted int    `json:"unposted"`
		Items    []repo `json:"items"`
	} `json:"data"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func GetRepository(limit int, posted bool) (*repositoryResponse, error) {

	url := os.Getenv("CONTENT_ALCHEMIST_URL") + "/think-root/api/get-repository/"

	payload := strings.NewReader(fmt.Sprintf(`{
		"limit": %d,
		"posted": %t
	}`, limit, posted))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("CONTENT_ALCHEMIST_BEARER"))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var response repositoryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &response, nil
}
