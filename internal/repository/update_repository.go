package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type updateResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type updatePostedRequest struct {
	URL    string `json:"url"`
	Posted bool   `json:"posted"`
}

func UpdateRepositoryPosted(url string, posted bool) (bool, error) {
	payload, err := json.Marshal(updatePostedRequest{URL: url, Posted: posted})
	if err != nil {
		return false, fmt.Errorf("error encoding request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPatch, updatePostedURL(), bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header = http.Header{
		"Accept":        {"*/*"},
		"Connection":    {"keep-alive"},
		"Content-Type":  {"application/json"},
		"Authorization": {authorizationHeader()},
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	var response updateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("error decoding response: %w", err)
	}

	return response.Status == "ok", nil
}
