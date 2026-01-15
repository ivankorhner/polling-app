package main

import (
	"context"
	"database/sql"
	"log/slog"

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

	// Connect to database
	db, err := sql.Open("pgx", cfg.DatabaseURL())
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to open database", slog.Any("error", err))
		return
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
		logger.LogAttrs(ctx, slog.LevelError, "failed to create user1", slog.Any("error", err))
		return
	}

	user2, err := client.User.Create().
		SetUsername("bob").
		SetEmail("bob@example.com").
		Save(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create user2", slog.Any("error", err))
		return
	}

	// Create polls
	poll1, err := client.Poll.Create().
		SetOwnerID(user1.ID).
		SetTitle("What's your favorite programming language?").
		Save(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create poll1", slog.Any("error", err))
		return
	}

	poll2, err := client.Poll.Create().
		SetOwnerID(user2.ID).
		SetTitle("Best web framework?").
		Save(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create poll2", slog.Any("error", err))
		return
	}

	// Create poll options for poll1
	opt1, err := client.PollOption.Create().
		SetPollID(poll1.ID).
		SetText("Go").
		SetVoteCount(0).
		Save(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create option1", slog.Any("error", err))
		return
	}

	opt2, err := client.PollOption.Create().
		SetPollID(poll1.ID).
		SetText("Rust").
		SetVoteCount(0).
		Save(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create option2", slog.Any("error", err))
		return
	}

	_, err = client.PollOption.Create().
		SetPollID(poll1.ID).
		SetText("Python").
		SetVoteCount(0).
		Save(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create option3", slog.Any("error", err))
		return
	}

	// Create poll options for poll2
	_, err = client.PollOption.Create().
		SetPollID(poll2.ID).
		SetText("React").
		SetVoteCount(0).
		Save(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create option4", slog.Any("error", err))
		return
	}

	_, err = client.PollOption.Create().
		SetPollID(poll2.ID).
		SetText("Vue").
		SetVoteCount(0).
		Save(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create option5", slog.Any("error", err))
		return
	}

	// Create votes
	_, err = client.Vote.Create().
		SetPollID(poll1.ID).
		SetOptionID(opt1.ID).
		SetUserID(user1.ID).
		Save(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create vote1", slog.Any("error", err))
		return
	}

	_, err = client.Vote.Create().
		SetPollID(poll1.ID).
		SetOptionID(opt2.ID).
		SetUserID(user2.ID).
		Save(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to create vote2", slog.Any("error", err))
		return
	}

	// Update vote counts
	if err := client.PollOption.UpdateOneID(opt1.ID).SetVoteCount(1).Exec(ctx); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to update option1 vote count", slog.Any("error", err))
		return
	}
	if err := client.PollOption.UpdateOneID(opt2.ID).SetVoteCount(1).Exec(ctx); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "failed to update option2 vote count", slog.Any("error", err))
		return
	}

	logger.Info("database seeded successfully",
		slog.Int("users", 2),
		slog.Int("polls", 2),
		slog.Int("options", 5),
		slog.Int("votes", 2),
	)
}
