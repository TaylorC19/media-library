package main

import (
	"log"
	"media-library/internal"
	"media-library/lib/db"
	// "github.com/gin-gonic/gin"
)

func main() {
	// Connect to MongoDB
	log.Println("Connecting to MongoDB")
	if err := db.Connect("mongodb://localhost:27017", "media_library"); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Disconnect()

	log.Println("Connected to MongoDB successfully")

	r := router.SetupRouter()
	// Listen and Server in 0.0.0.0:8080
	log.Println("Starting server on :8080")
	r.Run(":8080")
}
