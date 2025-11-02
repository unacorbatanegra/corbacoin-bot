package slack

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/unacorbatanegra/corbacoin-bot/config"
	"github.com/unacorbatanegra/corbacoin-bot/models"
)

// SendResponse sends a response to Slack using a response URL
func SendResponse(responseURL, text, responseType string) error {
	response := models.SlackResponse{
		Text:         text,
		ResponseType: responseType,
	}

	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}

	resp, err := http.Post(responseURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// SendErrorResponse sends an error message to Slack
func SendErrorResponse(responseURL, errorMessage, username string) error {
	message := fmt.Sprintf("❌ %s", errorMessage)
	if username != "" {
		message = fmt.Sprintf("❌ <@%s> %s", username, errorMessage)
	}
	return SendResponse(responseURL, message, "ephemeral")
}

// SendMessage sends a message to a Slack channel or thread
func SendMessage(channel, text, threadTS string) error {
	if config.SlackBotToken == "" {
		return fmt.Errorf("SLACK_BOT_TOKEN is not set")
	}

	message := models.SlackMessage{
		Channel:  channel,
		Text:     text,
		ThreadTS: threadTS,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SlackBotToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetUserInfo retrieves user information from Slack by user_id
func GetUserInfo(userID string) (*models.SlackUserInfo, error) {
	if config.SlackBotToken == "" {
		return nil, fmt.Errorf("SLACK_BOT_TOKEN is not set")
	}

	url := fmt.Sprintf("https://slack.com/api/users.info?user=%s", userID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SlackBotToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfoResponse models.SlackUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfoResponse); err != nil {
		return nil, err
	}

	if !userInfoResponse.Ok {
		return nil, fmt.Errorf("slack API error: %s", userInfoResponse.Error)
	}

	return &userInfoResponse.User, nil
}

// FindUserByUsername searches for a user in Slack by username and returns their user info
// This function searches through the workspace users to find a match by name or display name
func FindUserByUsername(username string) (*models.SlackUserInfo, error) {
	if config.SlackBotToken == "" {
		return nil, fmt.Errorf("SLACK_BOT_TOKEN is not set")
	}

	url := "https://slack.com/api/users.list"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SlackBotToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var usersListResponse models.SlackUsersListResponse
	if err := json.NewDecoder(resp.Body).Decode(&usersListResponse); err != nil {
		return nil, err
	}

	if !usersListResponse.Ok {
		return nil, fmt.Errorf("slack API error: %s", usersListResponse.Error)
	}

	// Search for user by username (case-insensitive)
	for _, user := range usersListResponse.Members {
		if user.Name == username || user.Profile.DisplayName == username {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user not found: %s", username)
}

// VerifySignature verifies that a request came from Slack
func VerifySignature(r *http.Request, body []byte) bool {
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	signature := r.Header.Get("X-Slack-Signature")

	if timestamp == "" || signature == "" {
		log.Println("Missing timestamp or signature headers")
		return false
	}

	// Check timestamp is within tolerance
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		log.Printf("Invalid timestamp: %v", err)
		return false
	}

	if time.Now().Unix()-ts > config.RequestTimestampTolerance {
		log.Println("Request timestamp is too old")
		return false
	}

	// Compute signature
	sigBaseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(config.SlackSigningSecret))
	mac.Write([]byte(sigBaseString))
	expectedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
