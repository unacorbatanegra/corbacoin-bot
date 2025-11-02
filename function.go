// Package corbacoin implements Slack bot handlers for Cloud Functions
package corbacoin

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/gin-gonic/gin"
	"github.com/unacorbatanegra/corbacoin-bot/config"
	"github.com/unacorbatanegra/corbacoin-bot/database"
	"github.com/unacorbatanegra/corbacoin-bot/handlers"
)

func init() {
	log.Println("Initializing Corbacoin Bot")


	// Load configuration from environment variables
	config.SlackBotToken = os.Getenv("SLACK_BOT_TOKEN")
	config.SlackSigningSecret = os.Getenv("SLACK_SIGNING_SECRET")

	if config.SlackSigningSecret == "" {
		log.Println("WARNING: SLACK_SIGNING_SECRET environment variable is not set")
	}

	// Initialize Firestore client
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("GCP_PROJECT")
	}
	if projectID == "" {
		log.Fatal("ERROR: GOOGLE_CLOUD_PROJECT or GCP_PROJECT environment variable must be set")
	}

	log.Printf("Initializing Firestore client for project: %s", projectID)
	var err error
	
	// Try to use the default database first
	database.Client, err = firestore.NewClientWithDatabase(ctx, projectID, config.FirestoreDatabase)
    if err != nil {
		log.Printf("Failed to create firestore client: %v", err)
		return
	}

	log.Println("Successfully connected to Firestore database")

	log.Println("Registering HTTP functions with Gin handlers wrapped for Cloud Functions")
	// Register HTTP functions with Gin handlers wrapped for Cloud Functions
	functions.HTTP("SlackCommandGo", ginToHTTPHandler(handlers.SlackCommand))
	functions.HTTP("SlackEventsGo", ginToHTTPHandler(handlers.SlackEvents))
	log.Println("HTTP functions registered")
}

// ginToHTTPHandler wraps a Gin handler to work with Cloud Functions
func ginToHTTPHandler(ginHandler gin.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a minimal Gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = r

		// Call the Gin handler
		ginHandler(c)
	}
}

