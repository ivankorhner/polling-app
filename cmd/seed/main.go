package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ivankorhner/polling-app/internal/config"
	"github.com/ivankorhner/polling-app/internal/ent"
	"github.com/ivankorhner/polling-app/internal/logging"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()
	logger := logging.NewLogger(slog.LevelInfo)

	if err := run(ctx, cfg, logger); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "seed failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	// Connect to database
	db, err := sql.Open("pgx", cfg.DatabaseURL())
	if err != nil {
		return err
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	logger.Info("seeding database with demo data")

	// Create users
	user1, err := client.User.Create().
		SetUsername("alice").
		SetEmail("alice@example.com").
		Save(ctx)
	if err != nil {
		return err
	}

	user2, err := client.User.Create().
		SetUsername("bob").
		SetEmail("bob@example.com").
		Save(ctx)
	if err != nil {
		return err
	}

	// Create polls
	poll1, err := client.Poll.Create().
		SetOwnerID(user1.ID).
		SetTitle("What's your favorite programming language?").
		Save(ctx)
	if err != nil {
		return err
	}

	poll2, err := client.Poll.Create().
		SetOwnerID(user2.ID).
		SetTitle("Best web framework?").
		Save(ctx)
	if err != nil {
		return err
	}

	// Create poll options for poll1
	opt1, err := client.PollOption.Create().
		SetPollID(poll1.ID).
		SetText("Go").
		Save(ctx)
	if err != nil {
		return err
	}

	opt2, err := client.PollOption.Create().
		SetPollID(poll1.ID).
		SetText("Rust").
		Save(ctx)
	if err != nil {
		return err
	}

	_, err = client.PollOption.Create().
		SetPollID(poll1.ID).
		SetText("Python").
		Save(ctx)
	if err != nil {
		return err
	}

	// Create poll options for poll2
	_, err = client.PollOption.Create().
		SetPollID(poll2.ID).
		SetText("React").
		Save(ctx)
	if err != nil {
		return err
	}

	_, err = client.PollOption.Create().
		SetPollID(poll2.ID).
		SetText("Vue").
		Save(ctx)
	if err != nil {
		return err
	}

	// Create votes (vote_count is now calculated dynamically)
	_, err = client.Vote.Create().
		SetPollID(poll1.ID).
		SetOptionID(opt1.ID).
		SetUserID(user1.ID).
		Save(ctx)
	if err != nil {
		return err
	}

	_, err = client.Vote.Create().
		SetPollID(poll1.ID).
		SetOptionID(opt2.ID).
		SetUserID(user2.ID).
		Save(ctx)
	if err != nil {
		return err
	}

	logger.Info("database seeded successfully",
		slog.Int("users", 2),
		slog.Int("polls", 2),
		slog.Int("options", 5),
		slog.Int("votes", 2),
	)

	return nil
}
