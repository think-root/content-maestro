package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Item is a single repository as returned by content-alchemist.
type Item struct {
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
		Items      []Item `json:"items"`
		Page       int    `json:"page"`
		PageSize   int    `json:"page_size"`
		TotalPages int    `json:"total_pages"`
		TotalItems int    `json:"total_items"`
	} `json:"data"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type getRepositoryRequest struct {
	Limit        int    `json:"limit,omitempty"`
	Posted       *bool  `json:"posted,omitempty"`
	SortOrder    string `json:"sort_order,omitempty"`
	SortBy       string `json:"sort_by,omitempty"`
	TextLanguage string `json:"text_language,omitempty"`
	URL          string `json:"url,omitempty"`
}

func GetRepository(limit int, posted bool, sort_order, sort_by string, textLanguage ...string) (*repositoryResponse, error) {
	var lang string
	if len(textLanguage) > 0 && textLanguage[0] != "" {
		lang = textLanguage[0]
	}

	return makeRepositoryRequest(getRepositoryRequest{
		Limit:        limit,
		Posted:       &posted,
		SortOrder:    sort_order,
		SortBy:       sort_by,
		TextLanguage: lang,
	})
}

// ErrRepositoryNotFound reports that content-alchemist has no repository for the
// requested identifier. It is a sentinel so HTTP callers can answer 404 for what
// is a client mistake - a typo'd or already deleted url - instead of 500, which
// would be indistinguishable from content-alchemist itself being down.
var ErrRepositoryNotFound = errors.New("repository not found")

// GetRepositoryByURL fetches one repository by its url regardless of its posted
// state. Used when a specific publication has to be re-sent to a connector, so
// the item must not be looked up through the publication queue.
func GetRepositoryByURL(url, textLanguage string) (*Item, error) {
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("repository url is required")
	}

	response, err := makeRepositoryRequest(getRepositoryRequest{
		URL:          url,
		TextLanguage: textLanguage,
	})
	if err != nil {
		return nil, err
	}

	if len(response.Data.Items) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrRepositoryNotFound, url)
	}

	item := response.Data.Items[0]

	// An older content-alchemist silently ignores the url filter and answers with
	// the head of the queue instead. Posting that item would publish the wrong
	// repository without any error, so refuse anything we did not ask for.
	if item.URL != url {
		return nil, fmt.Errorf("content-alchemist returned repository %s instead of %s", item.URL, url)
	}

	return &item, nil
}

// GetLatestPostedRepository returns the most recently published repository.
func GetLatestPostedRepository(textLanguage string) (*Item, error) {
	response, err := GetRepository(1, true, "DESC", "date_posted", textLanguage)
	if err != nil {
		return nil, err
	}

	if len(response.Data.Items) == 0 {
		return nil, fmt.Errorf("no published repositories found")
	}

	item := response.Data.Items[0]
	return &item, nil
}

func makeRepositoryRequest(payload getRepositoryRequest) (*repositoryResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error encoding request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, getRepositoryURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header = http.Header{
		"Accept":        {"*/*"},
		"Connection":    {"keep-alive"},
		"Content-Type":  {"application/json"},
		"Authorization": {authorizationHeader()},
	}

	resp, err := doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// A rejected request must surface as an error. Decoding it as a normal
	// payload leaves Items empty, which callers used to report as "no items
	// available" and hid the real cause.
	if resp.StatusCode == http.StatusNotFound {
		// content-alchemist answers 404 for an identifier it does not know, which is
		// a client mistake rather than a failure of either service.
		return nil, fmt.Errorf("%w: %s", ErrRepositoryNotFound, errorDetail(respBody))
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("content-alchemist error (status %d): %s", resp.StatusCode, errorDetail(respBody))
	}

	var response repositoryResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if response.Status == "error" {
		return nil, fmt.Errorf("content-alchemist error: %s", errorDetail(respBody))
	}

	return &response, nil
}

// errorDetail pulls the message out of a content-alchemist error envelope and
// falls back to the raw body, which is plain text for auth and rate-limit
// rejections.
func errorDetail(body []byte) string {
	var envelope struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Message != "" {
		return envelope.Message
	}

	detail := strings.TrimSpace(string(body))
	if detail == "" {
		return "empty response body"
	}
	if len(detail) > 300 {
		return detail[:300] + "…"
	}

	return detail
}
