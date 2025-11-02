package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/unacorbatanegra/corbacoin-bot/config"
	"github.com/unacorbatanegra/corbacoin-bot/models"
	"google.golang.org/api/iterator"
)

var (
	// Client is the Firestore client instance
	Client *firestore.Client
)

// GetUser retrieves a user from the database, creating a new one if it doesn't exist
func GetUser(ctx context.Context, username string) (*models.User, error) {
	userRef := Client.Collection("users").Doc(username)
	doc, err := userRef.Get(ctx)

	if err != nil {
		// User doesn't exist, create new user
		user := &models.User{
			Username: username,
			Coins:    config.InitialCoins,
		}
		_, err = userRef.Set(ctx, user)
		if err != nil {
			log.Printf("Error creating user %s: %v", username, err)
			return user, nil
		}
		log.Printf("User created: %s", username)
		return user, nil
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		log.Printf("Error parsing user data for %s: %v", username, err)
		return &models.User{Username: username, Coins: config.InitialCoins}, nil
	}

	return &user, nil
}

// UpdateCoins updates a user's coin balance by adding the specified amount
func UpdateCoins(ctx context.Context, username string, amount int) (int, error) {
	user, err := GetUser(ctx, username)
	if err != nil {
		return 0, err
	}

	newAmount := user.Coins + amount
	userRef := Client.Collection("users").Doc(username)

	_, err = userRef.Update(ctx, []firestore.Update{
		{Path: "coins", Value: newAmount},
	})

	if err != nil {
		log.Printf("Error updating coins for %s: %v", username, err)
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

