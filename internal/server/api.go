package server

import (
	"content-maestro/internal/models"
	"content-maestro/internal/store"
	"content-maestro/internal/validation"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-co-op/gocron"
)

type CronAPI struct {
	store      *store.Store
	schedulers map[string]*gocron.Scheduler
	jobs       models.JobRegistry
}

func NewCronAPI(store *store.Store, schedulers map[string]*gocron.Scheduler, jobs models.JobRegistry) *CronAPI {
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
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	successStr := r.URL.Query().Get("success")
	sortOrder := r.URL.Query().Get("sort")

	// Parse page parameter (default: 1)
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// Parse limit parameter (default: 20)
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// Parse sort parameter (default: "desc")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	offset := (page - 1) * limit

	var success *bool
	if successStr != "" {
		successVal, err := strconv.ParseBool(successStr)
		if err != nil {
			http.Error(w, "Invalid success parameter", http.StatusBadRequest)
			return
		}
		success = &successVal
	}

	// Get total count for pagination metadata
	totalCount, err := api.store.GetCronHistoryCount(cronName, success)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get paginated and sorted history
	history, err := api.store.GetCronHistory(cronName, success, offset, limit, sortOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate pagination metadata
	totalPages := (totalCount + limit - 1) / limit // Ceiling division
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
