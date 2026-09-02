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
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrInvalidRetryRequest marks a manual publication that was rejected because of
// its input, so callers can answer with 400 instead of 500. Its message says
// "request" rather than "retry": the dashboard shows it verbatim, and both manual
// endpoints return it.
var ErrInvalidRetryRequest = errors.New("invalid request")

// ErrPublishBusy reports that another publication holds publishMutex. Publishing
// on demand answers an HTTP request, so it refuses rather than queueing behind a
// cron run that can take minutes.
var ErrPublishBusy = errors.New("another publication is already running")

// ErrNoEnabledIntegrations reports that no integration is enabled, so a run that
// targets all of them has nothing to send to.
var ErrNoEnabledIntegrations = errors.New("no integration is enabled")

// ErrAlreadyPosted reports that the item left the publication queue between the
// dashboard rendering the row and the request arriving - a cron run in between,
// or a second dashboard. Publishing it again would duplicate the post.
var ErrAlreadyPosted = errors.New("repository is already published")

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

// publishMutex serialises everything that publishes an item: the message cron and
// both manual endpoints. Two of them publishing the same item concurrently - a
// double-clicked button, two open dashboards, or a manual run landing mid-cron -
// would post twice to the same connector.
var publishMutex sync.Mutex

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

// PublishOutcome is the per-connector result of a manual publication.
type PublishOutcome struct {
	APIName string `json:"api_name"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// PublishResult is the response of a manual publication.
type PublishResult struct {
	URL       string           `json:"url"`
	Status    int              `json:"status"`
	Message   string           `json:"message"`
	Succeeded []string         `json:"succeeded"`
	Failed    []string         `json:"failed"`
	Outcomes  []PublishOutcome `json:"outcomes"`
	// Posted reports that the item was marked as published and left the queue.
	Posted bool `json:"posted"`
	// PostedError explains why that marking failed. It is the one failure the
	// per-connector outcomes cannot show: the item went out but stayed in the
	// queue, so the scheduled run will publish it again.
	PostedError string `json:"posted_error,omitempty"`
}

// markPostedPolicy decides when a still-unposted item leaves the publication
// queue after a manual run.
type markPostedPolicy int

const (
	// markPostedWhenComplete is the retry policy: the item stays in the queue
	// until every requested connector has it. A retry exists to repair a partial
	// run, so marking it posted after another partial run re-creates the failure.
	markPostedWhenComplete markPostedPolicy = iota
	// markPostedOnAnySuccess is the cron policy: one connector receiving the item
	// is a publication. The connectors that failed are finished off from cron
	// history with the retry endpoint, which is why the run is recorded as manual.
	markPostedOnAnySuccess
)

// publishOptions describes one manual publication run.
type publishOptions struct {
	// url pins the repository before any connector is contacted. Empty falls back
	// to the most recently published one.
	url string
	// apiNames selects the connectors. Empty means every enabled one.
	apiNames []string
	// requireUnposted refuses an item that already left the queue, so a stale row
	// in one dashboard cannot re-publish what another one - or the cron - just sent.
	requireUnposted bool
	markPosted      markPostedPolicy
	// label prefixes the cron-history message: "Manual retry" / "Manual publish".
	label string
}

// RetryMessagePost re-sends one repository to the named APIs. It exists because a
// partially successful message run marks the item as posted, which drops it out
// of the publication queue even though some connectors never received it.
//
// When url is empty the most recently published repository is used. That is only
// a guess at what a partial run consumed, so callers that know the item - the
// dashboard reads it from the run details - should always pass it explicitly.
func RetryMessagePost(st store.StoreInterface, apiNames []string, url string) (*PublishResult, error) {
	requested, err := normalizeAPINames(apiNames)
	if err != nil {
		return nil, err
	}

	// A retry waits for whatever is publishing right now: it is a repair action, and
	// refusing it would leave the connectors that failed unrepaired.
	publishMutex.Lock()
	defer publishMutex.Unlock()

	return publishManually(st, publishOptions{
		url:        url,
		apiNames:   requested,
		markPosted: markPostedWhenComplete,
		label:      "Manual retry",
	})
}

// PublishNow publishes one repository to every enabled integration immediately,
// instead of waiting for its turn in the publication queue. It exists because a
// post that was lost after publication - a bad record deleted from the queue -
// cannot be recovered by promoting anything: the item is already gone from it.
//
// The item is marked as posted as soon as any integration accepts it, matching the
// message cron. The run is recorded as manual, so the connectors that failed are
// finished off from cron history with RetryMessagePost.
//
// Unlike the cron this does not revalidate the repository URL and delete dead
// ones: the cron picks blindly from the queue, while this publishes the row a
// human chose, and deleting that row mid-request would be astonishing.
func PublishNow(st store.StoreInterface, url string) (*PublishResult, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, fmt.Errorf("%w: url is required", ErrInvalidRetryRequest)
	}

	// Unlike a retry this refuses to queue: it answers a request with a spinner in
	// front of it, and a cron run ahead of it can hold the lock for minutes.
	if !publishMutex.TryLock() {
		return nil, ErrPublishBusy
	}
	defer publishMutex.Unlock()

	return publishManually(st, publishOptions{
		url:             url,
		requireUnposted: true,
		markPosted:      markPostedOnAnySuccess,
		label:           "Manual publish",
	})
}

// publishManually publishes one repository to a set of connectors and records the
// run in cron history under the message job. Shared by the retry and publish-now
// endpoints, which differ only in how they choose connectors and in when a
// still-unposted item is allowed to leave the queue.
//
// The caller must already hold publishMutex.
func publishManually(st store.StoreInterface, opts publishOptions) (*PublishResult, error) {
	apiConfigs := api.GetAPIConfigs()
	if apiConfigs == nil {
		return nil, fmt.Errorf("API configurations not loaded")
	}

	requested := opts.apiNames
	if len(requested) == 0 {
		requested = enabledAPINames(apiConfigs)
		if len(requested) == 0 {
			return nil, ErrNoEnabledIntegrations
		}
	}

	// Pin the target before contacting any connector so every API in this call
	// publishes the same repository.
	url := strings.TrimSpace(opts.url)
	itemPosted := false
	if url == "" {
		latest, err := repository.GetLatestPostedRepository("")
		if err != nil {
			return nil, fmt.Errorf("failed to resolve the latest published repository: %w", err)
		}
		url = latest.URL
		itemPosted = latest.Posted
	}

	if opts.requireUnposted {
		// One lookup before anything is published, deliberately kept even though the
		// connector loop fetches the item again a moment later: it is what makes an
		// item that already left the queue, and a url content-alchemist does not
		// know, a single clear error instead of a set of per-connector failures plus
		// a cron-history row for a run that could never have worked. The language is
		// empty because only the posted flag is read here.
		current, err := repository.GetRepositoryByURL(url, "")
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", url, err)
		}
		if current.Posted {
			return nil, fmt.Errorf("%w: %s", ErrAlreadyPosted, url)
		}
	}

	result := &PublishResult{URL: url}

	// One image per run, not one per connector: the cron shares a single image
	// across all of them, and each generation is a separate upstream fetch.
	imageName := ""
	defer func() {
		if imageName == "" {
			return
		}
		if err := os.Remove(imageName); err != nil && !os.IsNotExist(err) {
			log.Errorf("Failed to remove manual publication image %s: %v", imageName, err)
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
			log.Errorf("%s API error during %s: %v", apiName, opts.label, err)
			result.addFailure(apiName, err.Error())
		case resp.Success:
			log.Debugf("%s post created successfully during %s with language %s!", apiName, opts.label, textLanguage)
			result.addSuccess(apiName)
		default:
			log.Errorf("%s API request failed during %s (status %d): %s", apiName, opts.label, resp.StatusCode, string(resp.Body))
			result.addFailure(apiName, fmt.Sprintf("API request failed with status %d", resp.StatusCode))
		}
	}

	if !itemPosted && shouldMarkPosted(opts.markPosted, result) {
		if ok, err := repository.UpdateRepositoryPosted(url, true); err != nil || !ok {
			// The item published but stayed in the queue, so the scheduled run will
			// publish it again. Report it: this is the one failure the per-connector
			// outcomes cannot show.
			result.PostedError = postedErrorMessage(err)
			log.Errorf("Failed to update posted status for %s after %s: %v", url, opts.label, err)
		} else {
			result.Posted = true
		}
	}

	switch {
	case len(result.Succeeded) == 0:
		result.Status = 0
		result.Message = fmt.Sprintf("%s: nothing sent for %s. Errors: %s", opts.label, url, result.errorSummary())
	case len(result.Failed) > 0:
		result.Status = 2
		result.Message = fmt.Sprintf("%s: %s sent to: %s. Failed: %s. Errors: %s",
			opts.label, url, strings.Join(result.Succeeded, ", "), strings.Join(result.Failed, ", "), result.errorSummary())
	default:
		result.Status = 1
		result.Message = fmt.Sprintf("%s: %s sent to: %s", opts.label, url, strings.Join(result.Succeeded, ", "))
	}

	// Recorded under the message job so the dashboard's existing filters show it,
	// and so the failed connectors can be finished off with the retry endpoint.
	// No Pushover notification: a manual run is already being watched by whoever
	// triggered it.
	details := &models.MessageRunDetails{
		URL:    url,
		Sent:   result.Succeeded,
		Failed: result.Failed,
		Manual: true,
	}
	if err := st.LogCronExecutionDetails("message", result.Status, result.Message, details); err != nil {
		log.Errorf("Failed to log %s execution: %v", opts.label, err)
	}

	return result, nil
}

func (r *PublishResult) addSuccess(apiName string) {
	r.Succeeded = append(r.Succeeded, apiName)
	r.Outcomes = append(r.Outcomes, PublishOutcome{APIName: apiName, Success: true})
}

func (r *PublishResult) addFailure(apiName, message string) {
	r.Failed = append(r.Failed, apiName)
	r.Outcomes = append(r.Outcomes, PublishOutcome{APIName: apiName, Success: false, Error: message})
}

func (r *PublishResult) errorSummary() string {
	messages := make([]string, 0, len(r.Outcomes))
	for _, outcome := range r.Outcomes {
		if !outcome.Success {
			messages = append(messages, fmt.Sprintf("%s: %s", outcome.APIName, outcome.Error))
		}
	}
	return strings.Join(messages, "; ")
}

// shouldMarkPosted applies the caller's policy to a finished run.
func shouldMarkPosted(policy markPostedPolicy, result *PublishResult) bool {
	if len(result.Succeeded) == 0 {
		return false
	}
	return policy == markPostedOnAnySuccess || len(result.Failed) == 0
}

// postedErrorMessage describes a failed posted-status update. content-alchemist
// answering "not ok" without an error of its own still means the item stayed in
// the queue, so it needs wording too.
func postedErrorMessage(err error) string {
	if err != nil {
		return err.Error()
	}
	return "content-alchemist did not confirm the posted status"
}

// enabledAPINames lists the enabled connectors in a stable order. The config is a
// map, so without sorting the connectors would be contacted - and reported - in a
// different order on every call.
func enabledAPINames(configs *api.APIConfig) []string {
	names := make([]string, 0, len(configs.APIs))
	for name, endpoint := range configs.APIs {
		if endpoint.Enabled {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	return names
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
