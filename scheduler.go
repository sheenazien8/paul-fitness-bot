package main

import (
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

var scheduler *cron.Cron

type SchedulerCallbacks struct {
	SendWorkout       func(userID int64)
	SendWeightReminder func(userID int64)
	SendWeeklyReport  func(userID int64)
}

var callbacks SchedulerCallbacks

func InitScheduler(cbs SchedulerCallbacks) error {
	callbacks = cbs

	loc := time.FixedZone("WIB", 7*3600)
	scheduler = cron.New(cron.WithLocation(loc))

	scheduleAllUserJobs()

	scheduler.AddFunc("0 7 * * 1", func() {
		slog.Info("weekly report triggered")
		sendWeeklyReports()
	})

	scheduler.Start()
	slog.Info("scheduler started")
	return nil
}

func scheduleAllUserJobs() {
	users, err := GetAllActiveUsers()
	if err != nil {
		slog.Error("schedule: get active users failed", "error", err)
		return
	}

	for _, user := range users {
		scheduleUserCronJobs(user)
	}
}

func scheduleUserCronJobs(user User) {
	workoutDays := parseWorkoutDays(user.WorkoutDays)
	for _, day := range workoutDays {
		cronDow := day % 7
		expr := fmt.Sprintf("0 %d * * %d", user.NotificationHour, cronDow)
		userID := user.UserID
		slog.Info("scheduling workout", "user_id", userID, "cron", expr, "day", day)
		if _, err := scheduler.AddFunc(expr, func() {
			slog.Info("sending workout notification", "user_id", userID)
			callbacks.SendWorkout(userID)
		}); err != nil {
			slog.Error("failed to schedule workout", "error", err, "cron", expr)
		}
	}

	weightExpr := fmt.Sprintf("1 %d * * *", user.NotificationHour)
	userID := user.UserID
	slog.Info("scheduling weight reminder", "user_id", userID, "cron", weightExpr)
	if _, err := scheduler.AddFunc(weightExpr, func() {
		slog.Info("sending weight reminder", "user_id", userID)
		callbacks.SendWeightReminder(userID)
	}); err != nil {
		slog.Error("failed to schedule weight reminder", "error", err, "cron", weightExpr)
	}
}

func sendWeeklyReports() {
	users, err := GetAllActiveUsers()
	if err != nil {
		slog.Error("weekly report: get users failed", "error", err)
		return
	}
	for _, user := range users {
		callbacks.SendWeeklyReport(user.UserID)
	}
}

func StopScheduler() {
	if scheduler != nil {
		scheduler.Stop()
		slog.Info("scheduler stopped")
	}
}

func UpdateUserSchedule(userID int64) error {
	if scheduler != nil {
		scheduler.Stop()
	}

	loc := time.FixedZone("WIB", 7*3600)
	scheduler = cron.New(cron.WithLocation(loc))
	scheduleAllUserJobs()
	scheduler.AddFunc("0 7 * * 1", func() {
		slog.Info("weekly report triggered")
		sendWeeklyReports()
	})
	scheduler.Start()
	slog.Info("scheduler restarted for user", "user_id", userID)
	return nil
}

func FormatWeeklyReport(stats *WeeklyStats) string {
	var sb strings.Builder

	now := time.Now()
	weekStart := now.AddDate(0, 0, -7).Format("2 Jan")
	weekEnd := now.Format("2 Jan 2006")

	sb.WriteString(fmt.Sprintf("📊 *Laporan Mingguan — %s - %s*\n\n", weekStart, weekEnd))

	if stats.WorkoutCount == 0 {
		sb.WriteString("Hmm, minggu ini belum ada workout 😅\n")
		sb.WriteString("Tapi ga apa-apa, minggu depan bisa mulai lagi!\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("💪 Workout: %dx latihan\n", stats.WorkoutCount))
	sb.WriteString(fmt.Sprintf("⏱️ Total: %d menit\n", stats.TotalDuration))
	sb.WriteString(fmt.Sprintf("🔥 Kalori: %s kcal\n", formatNumber(stats.TotalCalories)))

	avgSat := math.Round(stats.AvgSatisfaction*10) / 10
	score := CalculateWorkoutScore(stats.TotalDuration, stats.TotalCalories, int(avgSat))
	sb.WriteString(fmt.Sprintf("⭐ Skor rata-rata: %.0f/100 %s\n\n", score, GetScoreEmoji(score)))

	if stats.WeightStart > 0 && stats.WeightEnd > 0 {
		change := stats.WeightChange
		changeStr := fmt.Sprintf("%.1f kg", math.Abs(change))
		if change < -0.05 {
			sb.WriteString(fmt.Sprintf("📈 Berat: %.1f → %.1f kg (-%s) 📉\n", stats.WeightStart, stats.WeightEnd, changeStr))
		} else if change > 0.05 {
			sb.WriteString(fmt.Sprintf("📈 Berat: %.1f → %.1f kg (+%s) 📈\n", stats.WeightStart, stats.WeightEnd, changeStr))
		} else {
			sb.WriteString(fmt.Sprintf("📈 Berat: %.1f kg (stabil) ➡️\n", stats.WeightEnd))
		}
	}

	if stats.StreakDays > 0 {
		sb.WriteString(fmt.Sprintf("🔥 Streak: %d hari!\n", stats.StreakDays))
	}

	if stats.BestDay != "" {
		sb.WriteString(fmt.Sprintf("\n🏆 Best Day: %s (skor: %.0f) 💪\n", stats.BestDay, stats.BestScore))
	}

	return sb.String()
}

func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}