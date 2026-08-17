package repository

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// client carries no timeout of its own: it is applied per request from the
// environment, so nothing about the content-alchemist connection depends on when
// this package was imported.
var client = &http.Client{}

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

// doRequest sends a request to content-alchemist under the configured timeout.
func doRequest(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(req.Context(), getContentAlchemistTimeout())
	defer cancel()

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	// The deferred cancel must not close the body before the caller reads it, so
	// the body is buffered and the connection released here.
	body, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	return resp, nil
}

// The content-alchemist location is resolved per call rather than at import
// time, so it does not depend on when the environment was populated and can be
// pointed at a stub in tests.
func contentAlchemistURL(path string) string {
	return strings.TrimRight(os.Getenv("CONTENT_ALCHEMIST_URL"), "/") + path
}

func getRepositoryURL() string {
	return contentAlchemistURL("/think-root/api/get-repository/")
}

func updatePostedURL() string {
	return contentAlchemistURL("/think-root/api/update-posted/")
}

func deleteRepositoryURL() string {
	return contentAlchemistURL("/think-root/api/delete-repository/")
}

func authorizationHeader() string {
	return "Bearer " + os.Getenv("CONTENT_ALCHEMIST_BEARER")
}
