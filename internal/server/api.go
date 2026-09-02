package server

import (
	apiExecutor "content-maestro/internal/api"
	"content-maestro/internal/models"
	"content-maestro/internal/schedule"
	"content-maestro/internal/store"
	"content-maestro/internal/validation"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-co-op/gocron"
)

type CronAPI struct {
	store      store.StoreInterface
	schedulers map[string]*gocron.Scheduler
	jobs       models.JobRegistry
}

func NewCronAPI(store store.StoreInterface, schedulers map[string]*gocron.Scheduler, jobs models.JobRegistry) *CronAPI {
	return &CronAPI{
		store:      store,
		schedulers: schedulers,
		jobs:       jobs,
	}
}

func (api *CronAPI) GetCrons(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	settings, err := api.store.GetAllCronSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (api *CronAPI) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cronName := strings.TrimPrefix(r.URL.Path, "/api/crons/")
	cronName = strings.TrimSuffix(cronName, "/schedule")

	scheduler, exists := api.schedulers[cronName]
	if !exists {
		http.Error(w, "Invalid cron name", http.StatusBadRequest)
		return
	}

	var req models.UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validation.ValidateCronExpression(req.Schedule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	setting, err := api.store.GetCronSetting(cronName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if setting == nil {
		http.Error(w, "Cron not found", http.StatusNotFound)
		return
	}

	setting.Schedule = req.Schedule
	_, err = api.store.UpdateCronSetting(setting.Name, setting.Schedule, setting.IsActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	scheduler.Stop()
	scheduler.Clear()

	if job, ok := api.jobs[cronName]; ok {
		scheduler.Cron(setting.Schedule).Do(job, scheduler)
		if setting.IsActive {
			scheduler.StartAsync()
		}
	}

	response := models.CronResponse{
		Status:  "success",
		Message: "Schedule updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *CronAPI) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cronName := strings.TrimPrefix(r.URL.Path, "/api/crons/")
	cronName = strings.TrimSuffix(cronName, "/status")

	scheduler, exists := api.schedulers[cronName]
	if !exists {
		http.Error(w, "Invalid cron name", http.StatusBadRequest)
		return
	}

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	setting, err := api.store.GetCronSetting(cronName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if setting == nil {
		http.Error(w, "Cron not found", http.StatusNotFound)
		return
	}

	setting.IsActive = req.IsActive
	updatedSetting, err := api.store.UpdateCronSetting(setting.Name, setting.Schedule, setting.IsActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	scheduler.Stop()
	scheduler.Clear()

	if updatedSetting.IsActive {
		if job, ok := api.jobs[cronName]; ok {
			scheduler.Cron(setting.Schedule).Do(job, scheduler)
			scheduler.StartAsync()
		}
	}

	response := models.CronResponse{
		Status:  "success",
		Message: "Status updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *CronAPI) UpdateCollectSettings(w http.ResponseWriter, r *http.Request) {
	var settings store.CollectSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if settings.MaxRepos < 1 {
		http.Error(w, "MaxRepos must be greater than 0", http.StatusBadRequest)
		return
	}
	if settings.Since == "" {
		http.Error(w, "Since cannot be empty", http.StatusBadRequest)
		return
	}
	if settings.SpokenLanguageCode == "" {
		http.Error(w, "SpokenLanguageCode cannot be empty", http.StatusBadRequest)
		return
	}

	if err := api.store.UpdateCollectSettings(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.CronResponse{
		Status:  "success",
		Message: "Collect settings updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *CronAPI) GetCollectSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := api.store.GetCollectSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (api *CronAPI) GetCronHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cronName := r.URL.Query().Get("name")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	statusStr := r.URL.Query().Get("status")
	sortOrder := r.URL.Query().Get("sort")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	offset := (page - 1) * limit

	var status *int
	if statusStr != "" {
		statusVal, err := strconv.Atoi(statusStr)
		if err != nil || statusVal < 0 || statusVal > 2 {
			http.Error(w, "Invalid status parameter: must be 0, 1, or 2", http.StatusBadRequest)
			return
		}
		status = &statusVal
	}

	startDate, endDate, err := validation.ParseDateRange(startDateStr, endDateStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	totalCount, err := api.store.GetCronHistoryCount(cronName, status, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	history, err := api.store.GetCronHistory(cronName, status, offset, limit, sortOrder, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := (totalCount + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	pagination := models.PaginationMetadata{
		TotalCount:  totalCount,
		CurrentPage: page,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	response := models.PaginatedCronHistoryResponse{
		Data:       history,
		Pagination: pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *CronAPI) GetPromptSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	settings, err := api.store.GetPromptSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (api *CronAPI) UpdatePromptSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings models.UpdatePromptSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validation.ValidatePromptSettings(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := api.store.UpdatePromptSettings(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.CronResponse{
		Status:  "success",
		Message: "Prompt settings updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *CronAPI) HandleCollectSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.GetCollectSettings(w, r)
	case http.MethodPut:
		api.UpdateCollectSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *CronAPI) HandlePromptSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.GetPromptSettings(w, r)
	case http.MethodPut:
		api.UpdatePromptSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *CronAPI) GetAPIConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := api.store.GetAllAPIConfigs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}

func (api *CronAPI) GetAPIConfig(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/api-configs/")

	config, err := api.store.GetAPIConfig(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if config == nil {
		http.Error(w, "API config not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (api *CronAPI) CreateAPIConfig(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAPIConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validation.ValidateAPIConfig(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config, err := api.store.CreateAPIConfig(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := apiExecutor.ReloadAPIConfigs(api.store); err != nil {
		http.Error(w, fmt.Sprintf("API config created but failed to reload: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(config)
}

func (api *CronAPI) UpdateAPIConfig(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/api-configs/")

	var req models.UpdateAPIConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validation.ValidateAPIConfigUpdate(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config, err := api.store.UpdateAPIConfig(name, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := apiExecutor.ReloadAPIConfigs(api.store); err != nil {
		http.Error(w, fmt.Sprintf("API config updated but failed to reload: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (api *CronAPI) DeleteAPIConfig(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/api-configs/")

	if err := api.store.DeleteAPIConfig(name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := apiExecutor.ReloadAPIConfigs(api.store); err != nil {
		http.Error(w, fmt.Sprintf("API config deleted but failed to reload: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.CronResponse{
		Status:  "success",
		Message: "API config deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RetryMessagePost re-sends an already published repository to the APIs that
// missed it. A partially successful message run marks the item as posted, so the
// cron alone can never recover the connectors that failed.
//
// The response is always 200 once the retry has run: per-API outcomes carry the
// individual failures, which are far more useful than a single status code.
func (api *CronAPI) RetryMessagePost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		return
	case http.MethodPost:
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RetryMessageRequest
	// The body is a short list of integration names and a url; anything larger is
	// not a request this endpoint should read into memory.
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8*1024)).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := schedule.RetryMessagePost(api.store, req.APIs, req.URL)
	if err != nil {
		if errors.Is(err, schedule.ErrInvalidRetryRequest) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// PublishMessageNow publishes a repository immediately instead of waiting for its
// turn in the publication queue. It exists because a post lost after publication -
// a bad record deleted from the queue - cannot be recovered by promoting anything.
//
// It targets every enabled integration and takes no API list: the dashboard's
// cached configuration must not decide what actually gets published.
//
// Like the retry endpoint this answers 200 once the run happened, however many
// integrations failed - the per-integration outcomes are the useful signal. The
// error statuses all describe why nothing was attempted at all.
func (api *CronAPI) PublishMessageNow(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		return
	case http.MethodPost:
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.PublishMessageRequest
	// The body is a single url; anything larger is not a request this endpoint
	// should read into memory.
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8*1024)).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := schedule.PublishNow(api.store, req.URL)
	if err != nil {
		switch {
		case errors.Is(err, schedule.ErrInvalidRetryRequest):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, schedule.ErrPublishBusy),
			errors.Is(err, schedule.ErrNoEnabledIntegrations),
			errors.Is(err, schedule.ErrAlreadyPosted):
			// Nothing about the request is wrong; the server is in a state that makes
			// publishing this item now impossible, and the message says which.
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (api *CronAPI) HandleAPIConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		return
	case http.MethodGet:
		api.GetAPIConfigs(w, r)
	case http.MethodPost:
		api.CreateAPIConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *CronAPI) HandleAPIConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		return
	case http.MethodGet:
		api.GetAPIConfig(w, r)
	case http.MethodPut:
		api.UpdateAPIConfig(w, r)
	case http.MethodDelete:
		api.DeleteAPIConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
