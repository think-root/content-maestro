package store

import (
	"content-maestro/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T) *SQLiteStore {
	store, err := NewSQLiteStore(":memory:")
	require.NoError(t, err)
	return store
}

func TestNewSQLiteStore(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	assert.NotNil(t, store)
	assert.NotNil(t, store.db)
}

func TestNewSQLiteStore_InvalidPath(t *testing.T) {
	_, err := NewSQLiteStore("/nonexistent/path/that/should/fail/db.sqlite")
	assert.Error(t, err)
}

func TestSQLiteStore_Close(t *testing.T) {
	store := setupTestStore(t)
	err := store.Close()
	assert.NoError(t, err)
}

func TestSQLiteStore_GetCronSetting(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Test getting existing default setting
	setting, err := store.GetCronSetting("collect")
	require.NoError(t, err)
	assert.NotNil(t, setting)
	assert.Equal(t, "collect", setting.Name)
	assert.Equal(t, "13 13 * * 6", setting.Schedule)
	assert.False(t, setting.IsActive) // Default cron jobs start disabled
}

func TestSQLiteStore_GetCronSetting_NotFound(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	setting, err := store.GetCronSetting("nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, setting)
}

func TestSQLiteStore_GetAllCronSettings(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	settings, err := store.GetAllCronSettings()
	require.NoError(t, err)
	assert.Len(t, settings, 2) // collect and message

	// Verify order
	assert.Equal(t, "collect", settings[0].Name)
	assert.Equal(t, "message", settings[1].Name)
}

func TestSQLiteStore_UpdateCronSetting(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Update existing setting
	updated, err := store.UpdateCronSetting("collect", "0 0 * * *", false)
	require.NoError(t, err)
	assert.Equal(t, "collect", updated.Name)
	assert.Equal(t, "0 0 * * *", updated.Schedule)
	assert.False(t, updated.IsActive)

	// Verify update persisted
	setting, err := store.GetCronSetting("collect")
	require.NoError(t, err)
	assert.Equal(t, "0 0 * * *", setting.Schedule)
	assert.False(t, setting.IsActive)
}

func TestSQLiteStore_UpdateCronSetting_NewSetting(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Create new setting
	created, err := store.UpdateCronSetting("new_cron", "30 8 * * 1-5", true)
	require.NoError(t, err)
	assert.Equal(t, "new_cron", created.Name)
	assert.Equal(t, "30 8 * * 1-5", created.Schedule)
	assert.True(t, created.IsActive)

	// Verify creation persisted
	setting, err := store.GetCronSetting("new_cron")
	require.NoError(t, err)
	assert.NotNil(t, setting)
}

func TestSQLiteStore_InitializeDefaultSettings(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Clear cron settings first (keep collect_settings to avoid id issue)
	_, err := store.db.Exec("DELETE FROM cron_settings")
	require.NoError(t, err)

	// Initialize defaults
	err = store.InitializeDefaultSettings()
	require.NoError(t, err)

	// Verify cron settings created
	settings, err := store.GetAllCronSettings()
	require.NoError(t, err)
	assert.Len(t, settings, 2)

	// Verify collect settings exist (were already initialized)
	collectSettings, err := store.GetCollectSettings()
	require.NoError(t, err)
	assert.NotNil(t, collectSettings)
}

func TestSQLiteStore_LogCronExecution(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	err := store.LogCronExecution("test_job", true, "Test output")
	require.NoError(t, err)

	// Verify log entry created
	history, err := store.GetCronHistory("test_job", nil, 0, 10, "desc", nil, nil)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "test_job", history[0].Name)
	assert.True(t, history[0].Success)
	assert.Equal(t, "Test output", history[0].Output)
}

func TestSQLiteStore_LogCronExecution_EmptyName(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	err := store.LogCronExecution("", true, "output")
	assert.Error(t, err)
}

func TestSQLiteStore_LogCronExecution_TruncateOutput(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Create output longer than 10000 chars
	longOutput := make([]byte, 15000)
	for i := range longOutput {
		longOutput[i] = 'a'
	}

	err := store.LogCronExecution("truncate_test", true, string(longOutput))
	require.NoError(t, err)

	history, err := store.GetCronHistory("truncate_test", nil, 0, 10, "desc", nil, nil)
	require.NoError(t, err)

	// Implementation truncates to (maxOutputLength-50) + marker
	// maxOutputLength = 10000, marker = "... [truncated due to length]" (30 chars)
	// Expected length: 9950 + 30 = 9980
	truncationMarker := "... [truncated due to length]"
	expectedLength := 10000 - 50 + len(truncationMarker) // 9980

	assert.Equal(t, expectedLength, len(history[0].Output),
		"truncated output should be exactly %d bytes (9950 content + %d marker)", expectedLength, len(truncationMarker))
	assert.True(t, len(history[0].Output) > 0 && history[0].Output[len(history[0].Output)-len(truncationMarker):] == truncationMarker,
		"truncated output should end with truncation marker: %q", truncationMarker)
}

func TestSQLiteStore_GetCronHistoryCount(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add some entries
	store.LogCronExecution("count_test", true, "success")
	store.LogCronExecution("count_test", false, "failure")
	store.LogCronExecution("other_job", true, "output")

	count, err := store.GetCronHistoryCount("count_test", nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Filter by success
	successTrue := true
	count, err = store.GetCronHistoryCount("count_test", &successTrue, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// All jobs
	count, err = store.GetCronHistoryCount("", nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestSQLiteStore_GetCronHistory_Pagination(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add entries
	for i := 0; i < 5; i++ {
		store.LogCronExecution("pagination_test", true, "output")
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get first page
	history, err := store.GetCronHistory("pagination_test", nil, 0, 2, "desc", nil, nil)
	require.NoError(t, err)
	assert.Len(t, history, 2)

	// Get second page
	history, err = store.GetCronHistory("pagination_test", nil, 2, 2, "desc", nil, nil)
	require.NoError(t, err)
	assert.Len(t, history, 2)

	// Get last page
	history, err = store.GetCronHistory("pagination_test", nil, 4, 2, "desc", nil, nil)
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestSQLiteStore_GetCronHistory_SortOrder(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Clear history
	store.db.Exec("DELETE FROM cron_history")

	// Add entries with different times
	store.LogCronExecution("sort_test", true, "first")
	time.Sleep(50 * time.Millisecond)
	store.LogCronExecution("sort_test", true, "second")

	// Test DESC order
	historyDesc, err := store.GetCronHistory("sort_test", nil, 0, 10, "desc", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "second", historyDesc[0].Output)

	// Test ASC order
	historyAsc, err := store.GetCronHistory("sort_test", nil, 0, 10, "asc", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "first", historyAsc[0].Output)
}

func TestSQLiteStore_GetCronHistory_DateFilter(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Clear history
	store.db.Exec("DELETE FROM cron_history")

	store.LogCronExecution("date_test", true, "output")

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	// Filter with date range including today
	history, err := store.GetCronHistory("date_test", nil, 0, 10, "desc", &yesterday, &tomorrow)
	require.NoError(t, err)
	assert.Len(t, history, 1)

	// Filter with date range in past
	pastStart := now.AddDate(0, 0, -10)
	pastEnd := now.AddDate(0, 0, -5)
	history, err = store.GetCronHistory("date_test", nil, 0, 10, "desc", &pastStart, &pastEnd)
	require.NoError(t, err)
	assert.Len(t, history, 0)
}

func TestSQLiteStore_GetCollectSettings(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	settings, err := store.GetCollectSettings()
	require.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, 5, settings.MaxRepos)
	assert.Equal(t, "daily", settings.Since)
	assert.Equal(t, "en", settings.SpokenLanguageCode)
}

func TestSQLiteStore_UpdateCollectSettings(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	newSettings := &CollectSettings{
		MaxRepos:           10,
		Since:              "weekly",
		SpokenLanguageCode: "uk",
	}

	err := store.UpdateCollectSettings(newSettings)
	require.NoError(t, err)

	// Verify update
	settings, err := store.GetCollectSettings()
	require.NoError(t, err)
	assert.Equal(t, 10, settings.MaxRepos)
	assert.Equal(t, "weekly", settings.Since)
	assert.Equal(t, "uk", settings.SpokenLanguageCode)
}

func TestSQLiteStore_GetPromptSettings(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	settings, err := store.GetPromptSettings()
	require.NoError(t, err)
	assert.NotNil(t, settings)
	assert.True(t, settings.UseDirectURL)
	assert.Equal(t, "openrouter", settings.LlmProvider)
	assert.Equal(t, 0.2, settings.Temperature)
	assert.Equal(t, "openai/gpt-4o-mini-search-preview", settings.Model)
	assert.Equal(t, "en,uk", settings.LlmOutputLanguage)
}

func TestSQLiteStore_UpdatePromptSettings(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	newProvider := "anthropic"
	newTemp := 0.5
	updateReq := &models.UpdatePromptSettingsRequest{
		LlmProvider: &newProvider,
		Temperature: &newTemp,
	}

	err := store.UpdatePromptSettings(updateReq)
	require.NoError(t, err)

	// Verify update
	settings, err := store.GetPromptSettings()
	require.NoError(t, err)
	assert.Equal(t, "anthropic", settings.LlmProvider)
	assert.Equal(t, 0.5, settings.Temperature)
	// Other fields should remain unchanged
	assert.True(t, settings.UseDirectURL)
}

func TestSQLiteStore_HasMigrationFlag(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Initially no migration flag
	hasMigration, err := store.HasMigrationFlag()
	require.NoError(t, err)
	assert.False(t, hasMigration)
}

func TestSQLiteStore_SetMigrationFlag(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	err := store.SetMigrationFlag()
	require.NoError(t, err)

	// Verify flag is set
	hasMigration, err := store.HasMigrationFlag()
	require.NoError(t, err)
	assert.True(t, hasMigration)
}

func TestSQLiteStore_InterfaceImplementation(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Verify SQLiteStore implements StoreInterface
	var _ StoreInterface = store
}
