package commands

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/unacorbatanegra/corbacoin-bot/config"
	"github.com/unacorbatanegra/corbacoin-bot/database"
	"github.com/unacorbatanegra/corbacoin-bot/models"
)

// HandleBalance returns the balance for a user
func HandleBalance(ctx context.Context, userID, username string) (string, error) {
	user, err := database.GetUser(ctx, userID, username)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("<@%s> has %d :corbacoin:", userID, user.Coins), nil
}

// ParseSendCommand parses a send command to extract recipient userID and amount
// Accepts both Slack mention format <@U12345678> and plain userID/username
func ParseSendCommand(text string) (recipientID string, amount int, ok bool) {
	log.Printf("ParseSendCommand input: '%s'", text)
	
	// Try to match Slack mention format first: <@U12345678> 100
	re := regexp.MustCompile(`<@([A-Z0-9]+)>\s+(\d+)`)
	matches := re.FindStringSubmatch(strings.TrimSpace(text))

	if len(matches) >= 3 {
		recipientID = matches[1]
		amount, err := strconv.Atoi(matches[2])
		if err != nil {
			return "", 0, false
		}
		log.Printf("ParseSendCommand matched mention format: recipientID=%s, amount=%d", recipientID, amount)
		return recipientID, amount, true
	}

	// Fallback to simple format: @username 100 or username 100
	re = regexp.MustCompile(`@?(\w+)\s+(\d+)`)
	matches = re.FindStringSubmatch(strings.TrimSpace(text))

	if len(matches) < 3 {
		log.Printf("ParseSendCommand: no match found")
		return "", 0, false
	}

	recipientID = matches[1]
	amount, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", 0, false
	}

	log.Printf("ParseSendCommand matched fallback format: recipientID=%s, amount=%d", recipientID, amount)
	return recipientID, amount, true
}

// HandleSend processes a send command to transfer coins between users
func HandleSend(ctx context.Context, senderID, senderName, recipientID, recipientName string, amount int) models.CommandResult {
	if amount <= 0 {
		return models.CommandResult{
			Success: false,
			Message: "Amount must be positive!",
		}
	}

	sender, err := database.GetUser(ctx, senderID, senderName)
	if err != nil {
		return models.CommandResult{
			Success: false,
			Message: "Error checking balance. Please try again.",
		}
	}

	if sender.Coins < amount {
		return models.CommandResult{
			Success: false,
			Message: fmt.Sprintf("Insufficient funds! You have %d :corbacoin:.", sender.Coins),
		}
	}

	// Get or create recipient user
	_, err = database.GetUser(ctx, recipientID, recipientName)
	if err != nil {
		return models.CommandResult{
			Success: false,
			Message: "Error finding recipient. Please try again.",
		}
	}

	// Perform transfer
	if _, err := database.UpdateCoins(ctx, senderID, -amount); err != nil {
		return models.CommandResult{
			Success: false,
			Message: "Error processing transfer. Please try again.",
		}
	}

	if _, err := database.UpdateCoins(ctx, recipientID, amount); err != nil {
		// Try to rollback
		database.UpdateCoins(ctx, senderID, amount)
		return models.CommandResult{
			Success: false,
			Message: "Error processing transfer. Please try again.",
		}
	}

	return models.CommandResult{
		Success: true,
		Message: fmt.Sprintf("<@%s> sent %d :corbacoin: to <@%s> :corbacoin:", senderID, amount, recipientID),
	}
}

// HandleLeaderboard returns the leaderboard of top users
func HandleLeaderboard(ctx context.Context) (string, error) {
	users, err := database.GetLeaderboard(ctx, config.LeaderboardLimit)
	if err != nil {
		return "", err
	}

	if len(users) == 0 {
		return "*Corbacoin Leaderboard* 🏆\nNo users found.", nil
	}

	var sb strings.Builder
	sb.WriteString("*Corbacoin Leaderboard* 🏆\n")
	for i, user := range users {
		sb.WriteString(fmt.Sprintf("%d. @%s: %d :corbacoin:\n", i+1, user.Username, user.Coins))
	}

	return sb.String(), nil
}

// GetHelpMessage returns the help message based on context
func GetHelpMessage(isAppMention bool) string {
	if isAppMention {
		return `*Corbacoin Bot Commands*

• ` + "`@CorbacoinBot balance`" + ` - Check your balance
• ` + "`@CorbacoinBot send @user amount`" + ` - Send corbacoins
• ` + "`@CorbacoinBot leaderboard`" + ` - View top 10 users
• ` + "`@CorbacoinBot help`" + ` - Show this message

You can use these in any channel or thread!`
	}

	return `*Corbacoin Slash Commands*

• ` + "`/balance`" + ` - Check your balance
• ` + "`/send @user amount`" + ` - Send corbacoins
• ` + "`/leaderboard`" + ` - View top 10 users`
}

