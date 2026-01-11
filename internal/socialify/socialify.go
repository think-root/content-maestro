package socialify

import (
	"content-maestro/internal/logger"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

var log = logger.NewLogger()
var SocialifyHTTPClient = &http.Client{}

type RetryConfig struct {
	MaxRetries    int
	RetryInterval time.Duration
}

var defaultConfig = RetryConfig{
	MaxRetries:    5,
	RetryInterval: 20 * time.Second,
}

var currentConfig = defaultConfig

func SetRetryConfig(config RetryConfig) {
	currentConfig = config
}

func ResetRetryConfig() {
	currentConfig = defaultConfig
}

func Socialify(usernameRepo string, outputPath string) error {
	log.Debug("Starting Socialify image parsing")

	var lastErr error
	for attempt := 1; attempt <= currentConfig.MaxRetries; attempt++ {
		err := trySocialify(usernameRepo, outputPath)
		if err == nil {
			log.Debug("Socialify image parsing finished")
			return nil
		}

		lastErr = err
		if attempt < currentConfig.MaxRetries {
			log.Errorf("Attempt %d failed: %v. Retrying in %s...", attempt, err, currentConfig.RetryInterval)
			time.Sleep(currentConfig.RetryInterval)
		}
	}

	log.Debugf("All %d attempts failed. Last error: %v", currentConfig.MaxRetries, lastErr)
	return lastErr
}

func trySocialify(usernameRepo string, outputPath string) error {
	patternsArray := []string{"Diagonal Stripes", "Charlie Brown", "Brick Wall", "Circuit Board", "Formal Invitation"}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomPattern := patternsArray[rng.Intn(len(patternsArray))]

	escapedPattern := url.QueryEscape(randomPattern)
	socialifyUrl := fmt.Sprintf(
		"https://socialify.git.ci/%s/png?description=0&font=Jost&forks=1&issues=1&language=1&name=1&owner=1&pattern=%s&pulls=1&stargazers=1&theme=Light",
		usernameRepo, escapedPattern,
	)

	req, err := http.NewRequest("GET", socialifyUrl, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) Content-Maestro/1.0")
	req.Header.Set("Accept", "image/png")

	response, err := SocialifyHTTPClient.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err := fmt.Errorf("received non-OK status code: %v", response.StatusCode)
		log.Error(err)
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		log.Error(err)
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
