package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"meto/internal"
)

func main() {
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	addr := os.Getenv("METO_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := server.NewServer()
	log.Printf("Starting server on %s", addr)
	if err := srv.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
