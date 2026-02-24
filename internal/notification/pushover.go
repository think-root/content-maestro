package notification

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const pushoverAPIURL = "https://api.pushover.net/1/messages.json"

func SendPushoverNotification(title, message string) {
	userKey := os.Getenv("PUSHOVER_USER_KEY")
	apiToken := os.Getenv("PUSHOVER_API_TOKEN")

	if userKey == "" || apiToken == "" {
		return
	}

	resp, err := http.PostForm(pushoverAPIURL, url.Values{
		"token":   {apiToken},
		"user":    {userKey},
		"title":   {title},
		"message": {message},
	})
	if err != nil {
		log.Errorf("Failed to send Pushover notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Pushover API returned status %d for notification: %s", resp.StatusCode, title)
	}
}

func NotifyCronResult(cronName string, status int, logMessage string) {
	if status == 1 {
		return
	}

	var statusLabel string
	switch status {
	case 0:
		statusLabel = "Failed"
	case 2:
		statusLabel = "Partial"
	default:
		statusLabel = fmt.Sprintf("Unknown(%d)", status)
	}

	title := fmt.Sprintf("content-maestro: %s %s", cronName, statusLabel)
	SendPushoverNotification(title, logMessage)
}
