package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Setup structured logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	slog.Info("starting workout bot")

	// Load .env
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables")
	}

	// Get bot token
	token := os.Getenv("WORKOUT_BOT_TOKEN")
	if token == "" {
		slog.Error("WORKOUT_BOT_TOKEN environment variable is required")
		os.Exit(1)
	}

	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/bot.db"
	}
	if err := InitDB(dbPath); err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer CloseDB()

	// Create default user (Sheena)
	defaultUser := &User{
		UserID:           491937914,
		Username:         "sheenazien",
		FirstName:        "Sheena",
		Weight:           72,
		Height:           167,
		TargetWeight:     65,
		WorkoutDays:      "1,2,4,5",
		NotificationHour: 7,
	}
	if err := CreateUser(defaultUser); err != nil {
		slog.Warn("default user may already exist", "error", err)
	}

	// Create bot
	app, err := NewBotApp(token)
	if err != nil {
		slog.Error("failed to create bot", "error", err)
		os.Exit(1)
	}

	// Initialize scheduler
	if err := InitScheduler(SchedulerCallbacks{
		SendWorkout: func(userID int64, dayOfWeek int) {
			app.SendWorkoutNotification(userID)
		},
		SendWeightReminder: func(userID int64) {
			app.SendWeightReminder(userID)
		},
	}); err != nil {
		slog.Error("failed to initialize scheduler", "error", err)
		os.Exit(1)
	}
	defer StopScheduler()

	// Setup updates channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := app.Bot.GetUpdatesChan(u)

	slog.Info("bot is running, waiting for messages...")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("shutting down...")
		StopScheduler()
		CloseDB()
		os.Exit(0)
	}()

	// Process updates
	for update := range updates {
		app.HandleUpdate(update)
	}
}