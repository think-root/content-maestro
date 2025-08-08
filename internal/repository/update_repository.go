package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type updateResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

var (
	client           *http.Client
	updatePostedUrl  string
	getRepositoryUrl string
	bearerToken      string
	once             sync.Once
)

func getContentAlchemistTimeout() time.Duration {
	timeoutStr := os.Getenv("CONTENT_ALCHEMIST_TIMEOUT")
	if timeoutStr == "" {
		return 30 * time.Second
	}

	timeoutSeconds, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return 30 * time.Second
	}

	return time.Duration(timeoutSeconds) * time.Second
}

func init() {
	once.Do(func() {
		updatePostedUrl = os.Getenv("CONTENT_ALCHEMIST_URL") + "/think-root/api/update-posted/"
		getRepositoryUrl = os.Getenv("CONTENT_ALCHEMIST_URL") + "/think-root/api/get-repository/"
		bearerToken = "Bearer " + os.Getenv("CONTENT_ALCHEMIST_BEARER")
		client = &http.Client{
			Timeout: getContentAlchemistTimeout(),
		}
	})
}

func UpdateRepositoryPosted(url string, posted bool) (bool, error) {
	payload := strings.NewReader(fmt.Sprintf(`{"url":"%s","posted":%t}`, url, posted))

	req, err := http.NewRequest(http.MethodPatch, updatePostedUrl, payload)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header = http.Header{
		"Accept":        {"*/*"},
		"Connection":    {"keep-alive"},
		"Content-Type":  {"application/json"},
		"Authorization": {bearerToken},
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
