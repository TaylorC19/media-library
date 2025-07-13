package main

import (
	"media-library/internal"
	// "github.com/gin-gonic/gin"
)

func main() {
	r := router.SetupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
