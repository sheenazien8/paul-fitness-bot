package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

var scheduler *cron.Cron

// SchedulerCallbacks holds functions to call from scheduled jobs
type SchedulerCallbacks struct {
	SendWorkout func(userID int64, dayOfWeek int)
	SendWeightReminder func(userID int64)
}

var callbacks SchedulerCallbacks

// InitScheduler sets up cron jobs for all registered users
func InitScheduler(cbs SchedulerCallbacks) error {
	callbacks = cbs

	scheduler = cron.New(cron.WithLocation(time.FixedZone("WIB", 7*3600)))

	// Add jobs for all users
	if err := scheduleUserJobs(); err != nil {
		return fmt.Errorf("schedule user jobs: %w", err)
	}

	scheduler.Start()
	slog.Info("scheduler started")
	return nil
}

// scheduleUserJobs adds cron jobs for workout notifications and weight reminders
func scheduleUserJobs() error {
	// For now, we schedule for the default user (Sheena)
	// This can be expanded to support multiple users
	userID := int64(491937914)

	user, err := GetUser(userID)
	if err != nil {
		slog.Warn("user not found for scheduling, will use defaults", "user_id", userID, "error", err)
		user = &User{
			UserID:           userID,
			WorkoutDays:      "1,2,4,5",
			NotificationHour: 7,
		}
	}

	// Parse workout days
	workoutDays := parseWorkoutDays(user.WorkoutDays)

	// Schedule workout notifications for each workout day
	for _, day := range workoutDays {
		// Convert day (1=Mon) to cron day-of-week (0=Sun in cron, but robfig uses 0=Sun)
		cronDow := day % 7 // 1=Mon -> 1, 2=Tue -> 2, 4=Thu -> 4, 5=Fri -> 5
		expr := fmt.Sprintf("0 %d * * %d", user.NotificationHour, cronDow)

		d := day // capture for closure
		slog.Info("scheduling workout", "user_id", userID, "cron", expr, "day", day)
		if _, err := scheduler.AddFunc(expr, func() {
			slog.Info("sending workout notification", "user_id", userID, "day", d)
			callbacks.SendWorkout(userID, d)
		}); err != nil {
			slog.Error("failed to schedule workout", "error", err, "cron", expr)
		}
	}

	// Schedule weight reminder every day at notification_hour + 1 minute
	weightExpr := fmt.Sprintf("1 %d * * *", user.NotificationHour)
	slog.Info("scheduling weight reminder", "user_id", userID, "cron", weightExpr)
	if _, err := scheduler.AddFunc(weightExpr, func() {
		slog.Info("sending weight reminder", "user_id", userID)
		callbacks.SendWeightReminder(userID)
	}); err != nil {
		slog.Error("failed to schedule weight reminder", "error", err, "cron", weightExpr)
	}

	return nil
}

// StopScheduler stops the cron scheduler
func StopScheduler() {
	if scheduler != nil {
		scheduler.Stop()
		slog.Info("scheduler stopped")
	}
}

// parseWorkoutDays parses "1,2,4,5" into []int{1,2,4,5}
func parseWorkoutDays(days string) []int {
	var result []int
	for _, s := range strings.Split(days, ",") {
		s = strings.TrimSpace(s)
		if n, err := strconv.Atoi(s); err == nil {
			result = append(result, n)
		}
	}
	return result
}

// UpdateUserSchedule reschedules jobs for a user
func UpdateUserSchedule(userID int64) error {
	// Remove existing jobs and re-add
	// Since cron doesn't easily support removing specific jobs,
	// we restart the scheduler
	if scheduler != nil {
		scheduler.Stop()
	}

	scheduler = cron.New(cron.WithLocation(time.FixedZone("WIB", 7*3600)))

	if err := scheduleUserJobs(); err != nil {
		return err
	}

	scheduler.Start()
	slog.Info("scheduler restarted for user", "user_id", userID)
	return nil
}