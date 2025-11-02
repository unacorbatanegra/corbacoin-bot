package commands

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/unacorbatanegra/corbacoin-bot/config"
	"github.com/unacorbatanegra/corbacoin-bot/database"
	"github.com/unacorbatanegra/corbacoin-bot/models"
)

// HandleBalance returns the balance for a user
func HandleBalance(ctx context.Context, username, userID string) (string, error) {
	user, err := database.GetUser(ctx, username)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("<@%s> has %d :corbacoin:", userID, user.Coins), nil
}

// ParseSendCommand parses a send command to extract recipient and amount
func ParseSendCommand(text string) (recipientName string, amount int, ok bool) {
	re := regexp.MustCompile(`@?(\w+)\s+(\d+)`)
	matches := re.FindStringSubmatch(strings.TrimSpace(text))

	if len(matches) < 3 {
		return "", 0, false
	}

	recipientName = matches[1]
	amount, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", 0, false
	}

	return recipientName, amount, true
}

// HandleSend processes a send command to transfer coins between users
func HandleSend(ctx context.Context, senderName, recipientName string, amount int) models.CommandResult {
	if amount <= 0 {
		return models.CommandResult{
			Success: false,
			Message: "Amount must be positive!",
		}
	}

	sender, err := database.GetUser(ctx, senderName)
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

	// Perform transfer
	if _, err := database.UpdateCoins(ctx, senderName, -amount); err != nil {
		return models.CommandResult{
			Success: false,
			Message: "Error processing transfer. Please try again.",
		}
	}

	if _, err := database.UpdateCoins(ctx, recipientName, amount); err != nil {
		// Try to rollback
		database.UpdateCoins(ctx, senderName, amount)
		return models.CommandResult{
			Success: false,
			Message: "Error processing transfer. Please try again.",
		}
	}

	return models.CommandResult{
		Success: true,
		Message: fmt.Sprintf("<@%s> sent %d :corbacoin: to <@%s> :corbacoin:", senderName, amount, recipientName),
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
		sb.WriteString(fmt.Sprintf("%d. %s: %d :corbacoin:\n", i+1, user.Username, user.Coins))
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

