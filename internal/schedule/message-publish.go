package schedule

import (
	"content-maestro/internal/api"
	"content-maestro/internal/models"
	"content-maestro/internal/repository"
	"content-maestro/internal/socialify"
	"content-maestro/internal/store"
	"content-maestro/internal/utils"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ErrInvalidRetryRequest marks a retry that was rejected because of its input,
// so callers can answer with 400 instead of 500.
var ErrInvalidRetryRequest = errors.New("invalid retry request")

// retrySocialifyConfig keeps image generation short: a manual retry answers an
// HTTP request, and the cron default (5 attempts, 20s apart) would block it for
// well over a minute.
var retrySocialifyConfig = socialify.RetryConfig{
	MaxRetries:    2,
	RetryInterval: 3 * time.Second,
}

const (
	imageDir      = "./tmp/gh_project_img"
	retryImageDir = imageDir + "/retry"
)

// retryMutex serialises manual retries. Two of them publishing the same item
// concurrently - a double-clicked button, or two open dashboards - would post
// twice to the same connector.
var retryMutex sync.Mutex

// imageURLPath turns a local image path into the path it is served under
// /images/, so an image kept in a subdirectory stays reachable.
func imageURLPath(imageName string) string {
	relative, err := filepath.Rel(filepath.Clean(imageDir), filepath.Clean(imageName))
	if err != nil || strings.HasPrefix(relative, "..") {
		return filepath.Base(imageName)
	}

	return filepath.ToSlash(relative)
}

// publishItem sends one repository to one configured API. Shared by the message
// cron and by manual retries so both build requests the same way.
func publishItem(apiName string, endpoint api.APIEndpoint, item repository.Item, imageName string) (*api.APIResponse, error) {
	var req api.RequestConfig

	commonFields := map[string]string{
		"text": item.Text,
		"url":  item.URL,
	}

	switch strings.ToLower(endpoint.ContentType) {
	case "multipart":
		req = api.RequestConfig{
			APIName:    apiName,
			FormFields: commonFields,
		}

		if endpoint.SocialifyImage && imageName != "" {
			req.FileFields = map[string]string{
				"image": imageName,
			}
		}
	case "json":
		req = api.RequestConfig{
			APIName:  apiName,
			JSONBody: map[string]any{"text": item.Text, "url": item.URL},
		}

		if endpoint.SocialifyImage && imageName != "" {
			publicURL := os.Getenv("PUBLIC_URL")
			if publicURL != "" {
				req.JSONBody["image_url"] = fmt.Sprintf("%s/images/%s", publicURL, imageURLPath(imageName))
			} else {
				log.Error("PUBLIC_URL not set, cannot generate image_url for API %s", apiName)
			}
		}
	default:
		req = api.RequestConfig{
			APIName:  apiName,
			JSONBody: map[string]any{"text": item.Text, "url": item.URL},
		}
	}

	return api.ExecuteRequest(req)
}

// RetryOutcome is the per-connector result of a manual retry.
type RetryOutcome struct {
	APIName string `json:"api_name"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// RetryResult is the response of a manual retry.
type RetryResult struct {
	URL       string         `json:"url"`
	Status    int            `json:"status"`
	Message   string         `json:"message"`
	Succeeded []string       `json:"succeeded"`
	Failed    []string       `json:"failed"`
	Outcomes  []RetryOutcome `json:"outcomes"`
}

// RetryMessagePost re-sends one repository to the named APIs. It exists because a
// partially successful message run marks the item as posted, which drops it out
// of the publication queue even though some connectors never received it.
//
// When url is empty the most recently published repository is used. That is only
// a guess at what a partial run consumed, so callers that know the item - the
// dashboard reads it from the run details - should always pass it explicitly.
func RetryMessagePost(st store.StoreInterface, apiNames []string, url string) (*RetryResult, error) {
	apiConfigs := api.GetAPIConfigs()
	if apiConfigs == nil {
		return nil, fmt.Errorf("API configurations not loaded")
	}

	requested, err := normalizeAPINames(apiNames)
	if err != nil {
		return nil, err
	}

	retryMutex.Lock()
	defer retryMutex.Unlock()

	// Pin the target before contacting any connector so every API in this call
	// publishes the same repository.
	url = strings.TrimSpace(url)
	itemPosted := false
	if url == "" {
		latest, err := repository.GetLatestPostedRepository("")
		if err != nil {
			return nil, fmt.Errorf("failed to resolve the latest published repository: %w", err)
		}
		url = latest.URL
		itemPosted = latest.Posted
	}

	result := &RetryResult{URL: url}

	// One image per retry, not one per connector: the cron shares a single image
	// across all of them, and each generation is a separate upstream fetch.
	imageName := ""
	defer func() {
		if imageName == "" {
			return
		}
		if err := os.Remove(imageName); err != nil && !os.IsNotExist(err) {
			log.Errorf("Failed to remove retry image %s: %v", imageName, err)
		}
	}()

	for _, apiName := range requested {
		endpoint, ok := apiConfigs.APIs[apiName]
		if !ok {
			result.addFailure(apiName, fmt.Sprintf("API %s is not configured", apiName))
			continue
		}
		if !endpoint.Enabled {
			result.addFailure(apiName, fmt.Sprintf("API %s is disabled", apiName))
			continue
		}

		textLanguage := endpoint.TextLanguage
		if textLanguage == "" {
			textLanguage = "en"
		}

		item, err := repository.GetRepositoryByURL(url, textLanguage)
		if err != nil {
			result.addFailure(apiName, fmt.Sprintf("failed to get repository (language %s): %v", textLanguage, err))
			continue
		}
		itemPosted = itemPosted || item.Posted

		if endpoint.SocialifyImage && imageName == "" {
			imageName, err = generateRetryImage(item.URL)
			if err != nil {
				result.addFailure(apiName, fmt.Sprintf("failed to prepare image: %v", err))
				continue
			}
		}

		resp, err := publishItem(apiName, endpoint, *item, imageName)

		switch {
		case err != nil:
			log.Errorf("%s API error during manual retry: %v", apiName, err)
			result.addFailure(apiName, err.Error())
		case resp.Success:
			log.Debugf("%s post created successfully during manual retry with language %s!", apiName, textLanguage)
			result.Succeeded = append(result.Succeeded, apiName)
			result.Outcomes = append(result.Outcomes, RetryOutcome{APIName: apiName, Success: true})
		default:
			log.Errorf("%s API request failed during manual retry (status %d): %s", apiName, resp.StatusCode, string(resp.Body))
			result.addFailure(apiName, fmt.Sprintf("API request failed with status %d", resp.StatusCode))
		}
	}

	// Marking an unposted item as posted while some connector still failed would
	// drop it out of the queue - the very failure this endpoint exists to repair -
	// so it is only marked once every requested connector has it.
	if !itemPosted && len(result.Succeeded) > 0 && len(result.Failed) == 0 {
		if _, err := repository.UpdateRepositoryPosted(url, true); err != nil {
			log.Errorf("Failed to update posted status for %s after manual retry: %v", url, err)
		}
	}

	switch {
	case len(result.Succeeded) == 0:
		result.Status = 0
		result.Message = fmt.Sprintf("Manual retry: nothing sent for %s. Errors: %s", url, result.errorSummary())
	case len(result.Failed) > 0:
		result.Status = 2
		result.Message = fmt.Sprintf("Manual retry: %s sent to: %s. Failed: %s. Errors: %s",
			url, strings.Join(result.Succeeded, ", "), strings.Join(result.Failed, ", "), result.errorSummary())
	default:
		result.Status = 1
		result.Message = fmt.Sprintf("Manual retry: %s sent to: %s", url, strings.Join(result.Succeeded, ", "))
	}

	// Recorded under the message job so the dashboard's existing filters show it.
	// No Pushover notification: a manual retry is already being watched by whoever
	// triggered it.
	details := &models.MessageRunDetails{
		URL:    url,
		Sent:   result.Succeeded,
		Failed: result.Failed,
		Manual: true,
	}
	if err := st.LogCronExecutionDetails("message", result.Status, result.Message, details); err != nil {
		log.Errorf("Failed to log manual retry execution: %v", err)
	}

	return result, nil
}

func (r *RetryResult) addFailure(apiName, message string) {
	r.Failed = append(r.Failed, apiName)
	r.Outcomes = append(r.Outcomes, RetryOutcome{APIName: apiName, Success: false, Error: message})
}

func (r *RetryResult) errorSummary() string {
	messages := make([]string, 0, len(r.Outcomes))
	for _, outcome := range r.Outcomes {
		if !outcome.Success {
			messages = append(messages, fmt.Sprintf("%s: %s", outcome.APIName, outcome.Error))
		}
	}
	return strings.Join(messages, "; ")
}

func normalizeAPINames(apiNames []string) ([]string, error) {
	seen := make(map[string]bool, len(apiNames))
	normalized := make([]string, 0, len(apiNames))

	for _, name := range apiNames {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		normalized = append(normalized, name)
	}

	if len(normalized) == 0 {
		return nil, fmt.Errorf("%w: at least one API name is required", ErrInvalidRetryRequest)
	}

	return normalized, nil
}

// generateRetryImage writes the image into a subdirectory of the image root, so
// the cron's cleanup - which only touches files - cannot delete it while the
// connector is still fetching it.
func generateRetryImage(repoURL string) (string, error) {
	if err := os.MkdirAll(retryImageDir, 0o777); err != nil {
		return "", fmt.Errorf("failed to create retry image directory: %w", err)
	}

	usernameRepo := strings.TrimPrefix(repoURL, "https://github.com/")
	imageName := fmt.Sprintf("%s/retry_%d.png", retryImageDir, time.Now().UnixNano())

	if err := socialify.SocialifyWithConfig(usernameRepo, imageName, retrySocialifyConfig); err != nil {
		log.Errorf("Socialify failed during manual retry: %v", err)
		if err := utils.CopyFile("./assets/banner.jpg", imageName); err != nil {
			// A failed generation can still have left a partial file behind.
			if removeErr := os.Remove(imageName); removeErr != nil && !os.IsNotExist(removeErr) {
				log.Errorf("Failed to remove partial retry image %s: %v", imageName, removeErr)
			}
			return "", fmt.Errorf("failed to copy fallback banner file: %w", err)
		}
	}

	return imageName, nil
}
