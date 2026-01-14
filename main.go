package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-playground/validator/v10"
)

func main() {
	cfg := LoadConfig()

	// Manual dependency injection
	validate := validator.New()
	pollService := NewPollService(store, validate)
	handlers := NewHandlers(pollService)

	// Register handlers
	http.HandleFunc("POST /polls", handlers.CreatePoll)
	http.HandleFunc("GET /polls", handlers.ListPolls)
	http.HandleFunc("GET /polls/{id}", handlers.GetPoll)
	http.HandleFunc("DELETE /polls/{id}", handlers.DeletePoll)
	http.HandleFunc("POST /polls/{id}/vote", handlers.Vote)

	log.Printf("Starting polling server on http://%s", cfg.Addr())
	log.Printf("Endpoints:")
	log.Printf("  POST   /polls          - Create a new poll")
	log.Printf("  GET    /polls          - List all polls")
	log.Printf("  GET    /polls/{id}     - Get a specific poll")
	log.Printf("  DELETE /polls/{id}     - Delete a poll")
	log.Printf("  POST   /polls/{id}/vote - Vote on a poll")

	if err := http.ListenAndServe(cfg.Addr(), nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func run(
	ctx context.Context,
	config *Config,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	handle := server.Add
}
