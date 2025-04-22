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

func Socialify(usernameRepo string) error {
	log.Debug("Starting Socialify image parsing")

	maxRetries := 5
	retryInterval := 20 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := trySocialify(usernameRepo)
		if err == nil {
			log.Debug("Socialify image parsing finished")
			return nil
		}

		lastErr = err
		if attempt < maxRetries {
			log.Errorf("Attempt %d failed: %v. Retrying in %s...", attempt, err, retryInterval)
			time.Sleep(retryInterval)
		}
	}

	log.Debugf("All %d attempts failed. Last error: %v", maxRetries, lastErr)
	return lastErr
}

func trySocialify(usernameRepo string) error {
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

	client := &http.Client{}
	response, err := client.Do(req)
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

	file, err := os.Create("./tmp/gh_project_img/image.png")
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
