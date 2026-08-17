package schedule

import (
	"content-maestro/internal/api"
	"content-maestro/internal/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// withMessageJobWorkdir gives the job an empty working directory containing the
// image folder it cleans up, so a test never writes into the repository.
func withMessageJobWorkdir(t *testing.T) {
	t.Helper()

	t.Chdir(t.TempDir())
	if err := os.MkdirAll(imageDir, 0o777); err != nil {
		t.Fatalf("failed to create image directory: %v", err)
	}
}

// queueStub serves the publication queue and can start rejecting requests after
// a given number of calls, which is how a re-fetch fails mid-run.
type queueStub struct {
	calls          int
	rejectAfter    int
	repositoryPath string
	repositoryURL  string
	validationCode int
}

func (q *queueStub) handler(t *testing.T) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		// The repository URL validation performs a HEAD request; answer it with the
		// configured status so the revalidation loop can be driven from a test.
		if r.Method == http.MethodHead {
			w.WriteHeader(q.validationCode)
			return
		}

		if r.Method == http.MethodDelete || r.Method == http.MethodPatch {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok","message":"done"}`))
			return
		}

		q.calls++
		if q.rejectAfter > 0 && q.calls > q.rejectAfter {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Too Many Requests"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"data": map[string]any{
				"items": []map[string]any{{
					"id": 1, "posted": false, "url": q.repositoryURL, "text": "text",
				}},
			},
		})
	}
}

// A failed re-fetch inside the revalidation loop used to leave the response nil
// and crash the job on the next read, which the deferred re-panic turned into a
// process exit.
func TestMessageJobSurvivesRejectedRefetch(t *testing.T) {
	stub := &queueStub{
		rejectAfter:    1,
		repositoryPath: "/dead/repo",
		validationCode: http.StatusNotFound,
	}
	server := httptest.NewServer(stub.handler(t))
	defer server.Close()
	stub.repositoryURL = server.URL + stub.repositoryPath
	withMessageJobWorkdir(t)

	t.Setenv("CONTENT_ALCHEMIST_URL", server.URL)
	t.Setenv("CONTENT_ALCHEMIST_BEARER", "test-token")

	st := &retryStore{configs: []models.APIConfigModel{{
		Name: "threads", URL: server.URL, Method: http.MethodPost,
		ContentType: "json", SuccessCode: http.StatusOK, Enabled: true,
		TextLanguage: "en",
	}}}
	if err := api.LoadAPIConfigs(st); err != nil {
		t.Fatalf("LoadAPIConfigs() error = %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MessageJob panicked: %v", r)
		}
	}()

	MessageJob(nil, st)

	if st.logCalls != 1 {
		t.Fatalf("cron history writes = %d, want 1", st.logCalls)
	}
	if st.loggedStatus != 0 {
		t.Errorf("status = %d, want 0", st.loggedStatus)
	}
	if !strings.Contains(st.loggedOutput, "failed to get next repository") {
		t.Errorf("output = %q, want the re-fetch failure reported", st.loggedOutput)
	}
	// The run consumed no item, so there is nothing to retry and no url to record.
	if st.loggedDetails == nil || len(st.loggedDetails.Failed) != 1 || st.loggedDetails.Failed[0] != "threads" {
		t.Fatalf("details = %+v, want threads recorded as failed", st.loggedDetails)
	}
	if st.loggedDetails.URL != "" {
		t.Errorf("details url = %q, want empty for a run that published nothing", st.loggedDetails.URL)
	}
}

// A run that cannot even read its configuration has nothing worth recording, so
// consumers can treat "has details" as "has something to retry".
func TestMessageJobRecordsNoDetailsWithoutConfigs(t *testing.T) {
	withMessageJobWorkdir(t)

	st := &retryStore{}
	if err := api.LoadAPIConfigs(st); err != nil {
		t.Fatalf("LoadAPIConfigs() error = %v", err)
	}

	MessageJob(nil, st)

	if st.logCalls != 1 {
		t.Fatalf("cron history writes = %d, want 1", st.logCalls)
	}
	if st.loggedDetails != nil {
		t.Errorf("details = %+v, want nil", st.loggedDetails)
	}
}

func TestMessageJobRecordsDetailsOnSuccess(t *testing.T) {
	stub := &queueStub{repositoryPath: "/live/repo", validationCode: http.StatusOK}
	server := httptest.NewServer(stub.handler(t))
	defer server.Close()
	repoURL := server.URL + stub.repositoryPath
	stub.repositoryURL = repoURL
	withMessageJobWorkdir(t)

	t.Setenv("CONTENT_ALCHEMIST_URL", server.URL)
	t.Setenv("CONTENT_ALCHEMIST_BEARER", "test-token")

	st := &retryStore{configs: []models.APIConfigModel{{
		Name: "threads", URL: server.URL, Method: http.MethodPost,
		ContentType: "json", SuccessCode: http.StatusOK, Enabled: true,
		TextLanguage: "en",
	}}}
	if err := api.LoadAPIConfigs(st); err != nil {
		t.Fatalf("LoadAPIConfigs() error = %v", err)
	}

	MessageJob(nil, st)

	if st.loggedStatus != 1 {
		t.Fatalf("status = %d, want 1 (output: %s)", st.loggedStatus, st.loggedOutput)
	}
	if st.loggedDetails == nil {
		t.Fatal("details = nil, want the published item recorded")
	}
	if st.loggedDetails.URL != repoURL {
		t.Errorf("details url = %q, want %q", st.loggedDetails.URL, repoURL)
	}
	if len(st.loggedDetails.Sent) != 1 || st.loggedDetails.Sent[0] != "threads" {
		t.Errorf("details sent = %v, want [threads]", st.loggedDetails.Sent)
	}
	if st.loggedDetails.Manual {
		t.Error("details manual = true, want false for a scheduled run")
	}
}
