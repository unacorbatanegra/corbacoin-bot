package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/unacorbatanegra/corbacoin-bot/commands"
	"github.com/unacorbatanegra/corbacoin-bot/models"
	"github.com/unacorbatanegra/corbacoin-bot/slack"
)

// SlackCommand handles Slack slash commands
func SlackCommand(c *gin.Context) {
	// Read body for signature verification
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	log.Println("SlackCommand body: " + string(body))

	// Verify Slack signature
	if !slack.VerifySignature(c.Request, body) {
		log.Println("Unauthorized: signature verification failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse form data
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	command := c.Request.FormValue("command")
	text := c.Request.FormValue("text")
	userID := c.Request.FormValue("user_id")
	userName := c.Request.FormValue("user_name")
	responseURL := c.Request.FormValue("response_url")

	// Send immediate acknowledgment
	acknowledgments := map[string]string{
		"/balance":     "⏳ Checking your balance...",
		"/send":        "⏳ Processing transfer...",
		"/leaderboard": "⏳ Loading leaderboard...",
	}

	ack := acknowledgments[command]
	if ack == "" {
		ack = "⏳ Processing..."
	}

	c.JSON(http.StatusOK, models.SlackResponse{
		Text:         ack,
		ResponseType: "ephemeral",
	})

	// Process command in background
	go func() {
		ctx := context.Background()

		switch command {
		case "/balance":
			message, err := commands.HandleBalance(ctx, userID, userName)
			if err != nil {
				slack.SendErrorResponse(responseURL, "An error occurred. Please try again later.", userID)
				return
			}
			slack.SendResponse(responseURL, message, "in_channel")

		case "/send":
			log.Println("Send command received: " + text)
			recipientIdentifier, amount, ok := commands.ParseSendCommand(text)
			if !ok {
				slack.SendErrorResponse(responseURL, "Usage: `/send @user amount`", userID)
				return
			}

			// Get recipient user info from Slack (works with both user_id and username)
			recipientInfo, err := slack.GetOrFindUser(recipientIdentifier)
			if err != nil {
				slack.SendErrorResponse(responseURL, fmt.Sprintf("Could not find user '%s'. %v", recipientIdentifier, err), userID)
				return
			}

			result := commands.HandleSend(ctx, userID, userName, recipientInfo.ID, recipientInfo.Name, amount)
			if result.Success {
				slack.SendResponse(responseURL, result.Message, "in_channel")
			} else {
				slack.SendErrorResponse(responseURL, result.Message, userID)
			}

		case "/leaderboard":
			message, err := commands.HandleLeaderboard(ctx)
			if err != nil {
				slack.SendErrorResponse(responseURL, "An error occurred. Please try again later.", userID)
				return
			}
			slack.SendResponse(responseURL, message, "in_channel")
		}
	}()
}

// SlackEvents handles Slack events
func SlackEvents(c *gin.Context) {
	log.Println("SlackEvents function called")
	// Read body for signature verification
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	log.Println("SlackEvents body: " + string(body))
	// Verify Slack signature
	if !slack.VerifySignature(c.Request, body) {
		log.Println("Unauthorized: signature verification failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var payload models.SlackEventPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	log.Printf("Slack events payload type: %s", payload.Type)

	// Handle URL verification challenge
	if payload.Type == "url_verification" {
		c.JSON(http.StatusOK, gin.H{"challenge": payload.Challenge})
		return
	}
	log.Printf("Payload: %+v", payload)

	

	// Acknowledge receipt
	c.Status(http.StatusOK)

	// Process event in background
	go func() {
		ctx := context.Background()
		event := payload.Event

		log.Printf("Event type: %s", event.Type)
		log.Println("Body: " + string(body))

		if event.Type == "app_mention" || event.Type == "message" {
			channel := event.Channel
			threadTS := event.ThreadTS
			if threadTS == "" {
				threadTS = event.TS
			}
			userName := event.User

			// Remove ONLY the first bot mention (not all mentions, as we need recipient mentions)
			re := regexp.MustCompile(`<@[A-Z0-9]+>\s*`)
			loc := re.FindStringIndex(event.Text)
			text := event.Text
			if loc != nil {
				// Remove only the first mention
				text = strings.TrimSpace(event.Text[:loc[0]] + event.Text[loc[1]:])
			}
			
			// For getting the command, we can lowercase the first word
			parts := strings.Fields(text)
			if len(parts) == 0 {
				return
			}

			command := strings.ToLower(parts[0])
			log.Printf("App mention received: command=%s, user=%s, channel=%s, fullText=%s", command, userName, channel, text)

			switch command {
			case "balance":
				message, err := commands.HandleBalance(ctx, userName, userName)
				if err != nil {
					log.Printf("Error handling balance: %v", err)
					return
				}
				slack.SendMessage(channel, message, threadTS)

			case "send":
				// Extract everything after the "send" command
				sendText := ""
				if len(parts) > 1 {
					// Find where "send" ends in the original text and take the rest
					idx := strings.Index(strings.ToLower(text), "send")
					if idx >= 0 {
						sendText = strings.TrimSpace(text[idx+4:]) // 4 is length of "send"
					}
				}
				
				recipientIdentifier, amount, ok := commands.ParseSendCommand(sendText)
				if !ok {
					slack.SendMessage(channel, "Usage: `@CorbacoinBot send @user amount`", threadTS)
					return
				}

				// Get recipient user info from Slack (works with both user_id and username)
				recipientInfo, err := slack.GetOrFindUser(recipientIdentifier)
				if err != nil {
					slack.SendMessage(channel, fmt.Sprintf("Could not find user '%s'. %v", recipientIdentifier, err), threadTS)
					return
				}

				result := commands.HandleSend(ctx, userName, userName, recipientInfo.ID, recipientInfo.Name, amount)
				slack.SendMessage(channel, result.Message, threadTS)

			case "leaderboard":
				message, err := commands.HandleLeaderboard(ctx)
				if err != nil {
					log.Printf("Error handling leaderboard: %v", err)
					return
				}
				slack.SendMessage(channel, message, threadTS)

			case "help":
				message := commands.GetHelpMessage(true)
				slack.SendMessage(channel, message, threadTS)

			default:
				slack.SendMessage(channel, "Unknown command. Try `@CorbacoinBot help` for available commands.", threadTS)
			}
		}
	}()
}
