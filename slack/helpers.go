package slack

import (
	"fmt"

	"github.com/unacorbatanegra/corbacoin-bot/models"
)

// GetOrFindUser returns user info by checking if the identifier is a user_id or username
// If it's a user_id (starts with U and is at least 9 chars), it fetches info directly
// If it's a username, it searches for the user in the workspace
func GetOrFindUser(identifier string) (*models.SlackUserInfo, error) {
	// Check if it looks like a Slack user ID (starts with U and is at least 9 characters)
	if len(identifier) > 0 && identifier[0] == 'U' && len(identifier) >= 9 {
		// It's likely a user_id, get user info directly
		return GetUserInfo(identifier)
	}
	
	// It's a username, search for it
	userInfo, err := FindUserByUsername(identifier)
	if err != nil {
		return nil, fmt.Errorf("user '%s' not found in Slack workspace", identifier)
	}
	
	return userInfo, nil
}

