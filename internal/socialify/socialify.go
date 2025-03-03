package socialify

import (
	"content-maestro/internal/logger"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var log = logger.NewLogger()

func Socialify(usernameRepo string) error {
	log.Debug("Starting Socialify image parsing")

	maxRetries := 3
	retryInterval := 10 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := trySocialify(usernameRepo)
		if err == nil {
			log.Error("Socialify image parsing finished")
			return nil
		}

		lastErr = err
		if attempt < maxRetries {
			log.Errorln("Attempt %d failed: %v. Retrying in %s...", attempt, err, retryInterval)
			time.Sleep(retryInterval)
		}
	}

	log.Debug("All %d attempts failed. Last error: %v", maxRetries, lastErr)
	return lastErr
}

func trySocialify(usernameRepo string) error {
	patterns := []string{"Diagonal Stripes", "Charlie Brown", "Brick Wall", "Circuit Board", "Formal Invitation", "Signal"}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomPattern := patterns[rng.Intn(len(patterns))]
	socialifyUrl := fmt.Sprintf("https://socialify.git.ci/%s/png?description=0&font=Jost&forks=1&issues=1&language=1&name=1&owner=1&pattern=%s&pulls=1&stargazers=1&theme=Light", usernameRepo, randomPattern)

	response, err := http.Get(socialifyUrl)
	if err != nil {
		log.Error(err)
		return err
	}
	defer response.Body.Close()

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

	if response.StatusCode != 200 {
		return fmt.Errorf("error code: %v", response.StatusCode)
	}

	return nil
}
