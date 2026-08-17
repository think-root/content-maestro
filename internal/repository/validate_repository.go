package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func ValidateRepositoryURL(url string) (int, error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}

	httpClient := &http.Client{
		Timeout: getContentAlchemistTimeout(),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

type deleteResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func DeleteRepository(url string) (bool, error) {
	payload, err := json.Marshal(map[string]string{"url": url})
	if err != nil {
		return false, fmt.Errorf("error encoding request: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, deleteRepositoryURL(), bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header = http.Header{
		"Accept":        {"*/*"},
		"Connection":    {"keep-alive"},
		"Content-Type":  {"application/json"},
		"Authorization": {authorizationHeader()},
	}

	resp, err := doRequest(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	var response deleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("error decoding response: %w", err)
	}

	return response.Status == "ok", nil
}
