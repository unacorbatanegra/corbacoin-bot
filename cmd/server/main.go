// Package main provides local development server for Corbacoin Bot
package main

import (
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	
	// Import corbacoin package to trigger init() and register functions
	_ "github.com/unacorbatanegra/corbacoin-bot"
)

func main() {
	// Use PORT environment variable, or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting Functions Framework server on port %s", port)
	
	// Start the Functions Framework server
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v", err)
	}
}

