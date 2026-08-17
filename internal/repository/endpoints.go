package repository

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	client     *http.Client
	clientOnce sync.Once
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
	clientOnce.Do(func() {
		client = &http.Client{
			Timeout: getContentAlchemistTimeout(),
		}
	})
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
