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

type SchedulerCallbacks struct {
	SendWorkout      func(userID int64, dayOfWeek int)
	SendWeightReminder func(userID int64)
}

var callbacks SchedulerCallbacks

func InitScheduler(cbs SchedulerCallbacks) error {
	callbacks = cbs

	scheduler = cron.New(cron.WithLocation(time.FixedZone("WIB", 7*3600)))

	if err := scheduleUserJobs(); err != nil {
		return fmt.Errorf("schedule user jobs: %w", err)
	}

	scheduler.Start()
	slog.Info("scheduler started")
	return nil
}

func scheduleUserJobs() error {
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

	workoutDays := parseWorkoutDays(user.WorkoutDays)

	for _, day := range workoutDays {
		cronDow := day % 7
		expr := fmt.Sprintf("0 %d * * %d", user.NotificationHour, cronDow)

		d := day
		slog.Info("scheduling workout", "user_id", userID, "cron", expr, "day", day)
		if _, err := scheduler.AddFunc(expr, func() {
			slog.Info("sending workout notification", "user_id", userID, "day", d)
			callbacks.SendWorkout(userID, d)
		}); err != nil {
			slog.Error("failed to schedule workout", "error", err, "cron", expr)
		}
	}

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

func StopScheduler() {
	if scheduler != nil {
		scheduler.Stop()
		slog.Info("scheduler stopped")
	}
}

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

func UpdateUserSchedule(userID int64) error {
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