package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/unacorbatanegra/corbacoin-bot/config"
	"github.com/unacorbatanegra/corbacoin-bot/models"
	"github.com/unacorbatanegra/corbacoin-bot/slack"
	"google.golang.org/api/iterator"
)

var (
	// Client is the Firestore client instance
	Client *firestore.Client
)

// GetUser retrieves a user from the database, creating a new one if it doesn't exist
func GetUser(ctx context.Context, userID, username string) (*models.User, error) {
	userRef := Client.Collection("users").Doc(userID)
	doc, err := userRef.Get(ctx)

	if err != nil {
		// User doesn't exist in database, fetch info from Slack
		actualUsername := username
		if actualUsername == "" {
			// If username is not provided, fetch it from Slack
			userInfo, slackErr := slack.GetUserInfo(userID)
			if slackErr != nil {
				log.Printf("Error fetching user info from Slack for %s: %v", userID, slackErr)
				// Fallback to userID as username if Slack fetch fails
				actualUsername = userID
			} else {
				actualUsername = userInfo.Name
			}
		}

		// Create new user
		user := &models.User{
			UserID:   userID,
			Username: actualUsername,
			Coins:    config.InitialCoins,
		}
		_, err = userRef.Set(ctx, user)
		if err != nil {
			log.Printf("Error creating user %s (%s): %v", actualUsername, userID, err)
			return user, nil
		}
		log.Printf("User created: %s (%s)", actualUsername, userID)
		return user, nil
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		log.Printf("Error parsing user data for %s (%s): %v", username, userID, err)
		return &models.User{UserID: userID, Username: username, Coins: config.InitialCoins}, nil
	}

	return &user, nil
}

// UpdateCoins updates a user's coin balance by adding the specified amount
func UpdateCoins(ctx context.Context, userID string, amount int) (int, error) {
	userRef := Client.Collection("users").Doc(userID)
	doc, err := userRef.Get(ctx)
	if err != nil {
		log.Printf("Error getting user %s: %v", userID, err)
		return 0, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		log.Printf("Error parsing user data for %s: %v", userID, err)
		return 0, err
	}

	newAmount := user.Coins + amount
	_, err = userRef.Update(ctx, []firestore.Update{
		{Path: "coins", Value: newAmount},
	})

	if err != nil {
		log.Printf("Error updating coins for %s: %v", userID, err)
		return 0, err
	}

	return newAmount, nil
}

// GetLeaderboard retrieves the top users by coin balance
func GetLeaderboard(ctx context.Context, limit int) ([]models.User, error) {
	if limit <= 0 {
		limit = config.LeaderboardLimit
	}

	query := Client.Collection("users").
		OrderBy("coins", firestore.Desc).
		Limit(limit)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var users []models.User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error iterating leaderboard: %v", err)
			return users, err
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			log.Printf("Error parsing user data: %v", err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

