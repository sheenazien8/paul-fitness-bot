package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func InitDB(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create db directory: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		slog.Warn("failed to enable WAL mode", "error", err)
	}

	if err := migrate(); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	slog.Info("database initialized", "path", dbPath)
	return nil
}

func migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY,
			username TEXT,
			first_name TEXT,
			weight REAL DEFAULT 72,
			height REAL DEFAULT 167,
			target_weight REAL DEFAULT 65,
			workout_days TEXT DEFAULT '1,2,4,5',
			notification_hour INTEGER DEFAULT 7,
			streak INTEGER DEFAULT 0,
			last_workout_date TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS workouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			day_of_week INTEGER,
			workout_type TEXT,
			exercises TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS workout_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			workout_id INTEGER,
			duration_minutes INTEGER,
			calories INTEGER,
			satisfaction INTEGER,
			score REAL,
			logged_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id),
			FOREIGN KEY (workout_id) REFERENCES workouts(id)
		)`,
		`CREATE TABLE IF NOT EXISTS weight_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			weight REAL,
			bmi REAL,
			logged_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_preferences (
			user_id INTEGER PRIMARY KEY,
			goal TEXT DEFAULT 'diet',
			experience_level TEXT DEFAULT 'beginner',
			has_dumbbell INTEGER DEFAULT 1,
			has_resistance_band INTEGER DEFAULT 1,
			has_pullup_bar INTEGER DEFAULT 1,
			onboarding_done INTEGER DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_sessions (
			user_id INTEGER PRIMARY KEY,
			state TEXT DEFAULT 'idle',
			workout_id INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS chat_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_history_user ON chat_history(user_id, id)`,
		`CREATE TABLE IF NOT EXISTS mood_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			mood TEXT NOT NULL,
			energy INTEGER NOT NULL,
			logged_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

func GetDB() *sql.DB {
	return db
}

func CreateUser(user *User) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO users (user_id, username, first_name, weight, height, target_weight, workout_days, notification_hour)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.UserID, user.Username, user.FirstName, user.Weight, user.Height, user.TargetWeight, user.WorkoutDays, user.NotificationHour,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func GetUser(userID int64) (*User, error) {
	u := &User{}
	var lastWorkoutDate sql.NullString
	err := db.QueryRow(
		`SELECT user_id, username, first_name, weight, height, target_weight, workout_days, notification_hour, streak, last_workout_date, created_at, updated_at
		 FROM users WHERE user_id = ?`, userID,
	).Scan(&u.UserID, &u.Username, &u.FirstName, &u.Weight, &u.Height, &u.TargetWeight, &u.WorkoutDays, &u.NotificationHour, &u.Streak, &lastWorkoutDate, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if lastWorkoutDate.Valid {
		u.LastWorkoutDate = lastWorkoutDate.String
	}
	return u, nil
}

func UpdateUserWeight(userID int64, weight float64) error {
	_, err := db.Exec(`UPDATE users SET weight = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?`, weight, userID)
	return err
}

func UpdateUserStreak(userID int64, streak int, lastWorkoutDate string) error {
	_, err := db.Exec(`UPDATE users SET streak = ?, last_workout_date = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?`, streak, lastWorkoutDate, userID)
	return err
}

func UpdateUserSettings(userID int64, workoutDays string, notificationHour int) error {
	_, err := db.Exec(`UPDATE users SET workout_days = ?, notification_hour = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?`, workoutDays, notificationHour, userID)
	return err
}

func UpdateUserProfile(userID int64, weight, height, targetWeight float64) error {
	_, err := db.Exec(`UPDATE users SET weight = ?, height = ?, target_weight = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?`, weight, height, targetWeight, userID)
	return err
}

func SaveWorkout(userID int64, dayOfWeek int, workoutType string, exercises []Exercise) (int64, error) {
	exJSON, err := json.Marshal(exercises)
	if err != nil {
		return 0, fmt.Errorf("marshal exercises: %w", err)
	}

	result, err := db.Exec(
		`INSERT INTO workouts (user_id, day_of_week, workout_type, exercises) VALUES (?, ?, ?, ?)`,
		userID, dayOfWeek, workoutType, string(exJSON),
	)
	if err != nil {
		return 0, fmt.Errorf("save workout: %w", err)
	}

	id, _ := result.LastInsertId()
	return id, nil
}

func GetLatestWorkout(userID int64) (*Workout, error) {
	w := &Workout{}
	err := db.QueryRow(
		`SELECT id, user_id, day_of_week, workout_type, exercises, created_at
		 FROM workouts WHERE user_id = ? ORDER BY id DESC LIMIT 1`, userID,
	).Scan(&w.ID, &w.UserID, &w.DayOfWeek, &w.WorkoutType, &w.Exercises, &w.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get latest workout: %w", err)
	}
	return w, nil
}

func GetTodaysWorkout(userID int64, dayOfWeek int) (*Workout, error) {
	w := &Workout{}
	today := time.Now().Format("2006-01-02")
	err := db.QueryRow(
		`SELECT id, user_id, day_of_week, workout_type, exercises, created_at
		 FROM workouts WHERE user_id = ? AND date(created_at) = ? ORDER BY id DESC LIMIT 1`, userID, today,
	).Scan(&w.ID, &w.UserID, &w.DayOfWeek, &w.WorkoutType, &w.Exercises, &w.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get today workout: %w", err)
	}
	return w, nil
}

func SaveWorkoutLog(log *WorkoutLog) error {
	_, err := db.Exec(
		`INSERT INTO workout_logs (user_id, workout_id, duration_minutes, calories, satisfaction, score) VALUES (?, ?, ?, ?, ?, ?)`,
		log.UserID, log.WorkoutID, log.DurationMinutes, log.Calories, log.Satisfaction, log.Score,
	)
	return err
}

func GetWeeklyWorkoutLogs(userID int64) ([]WorkoutLog, error) {
	rows, err := db.Query(
		`SELECT id, user_id, workout_id, duration_minutes, calories, satisfaction, score, logged_at
		 FROM workout_logs WHERE user_id = ? AND logged_at >= date('now', '-7 days') ORDER BY logged_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []WorkoutLog
	for rows.Next() {
		var l WorkoutLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.WorkoutID, &l.DurationMinutes, &l.Calories, &l.Satisfaction, &l.Score, &l.LoggedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func GetRecentWorkoutLogs(userID int64, limit int) ([]WorkoutLog, error) {
	rows, err := db.Query(
		`SELECT id, user_id, workout_id, duration_minutes, calories, satisfaction, score, logged_at
		 FROM workout_logs WHERE user_id = ? ORDER BY logged_at DESC LIMIT ?`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []WorkoutLog
	for rows.Next() {
		var l WorkoutLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.WorkoutID, &l.DurationMinutes, &l.Calories, &l.Satisfaction, &l.Score, &l.LoggedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func SaveWeightLog(userID int64, weight, bmi float64) error {
	_, err := db.Exec(
		`INSERT INTO weight_logs (user_id, weight, bmi) VALUES (?, ?, ?)`, userID, weight, bmi,
	)
	if err != nil {
		return err
	}
	return UpdateUserWeight(userID, weight)
}

func GetLatestWeightLog(userID int64) (*WeightLog, error) {
	w := &WeightLog{}
	err := db.QueryRow(
		`SELECT id, user_id, weight, bmi, logged_at FROM weight_logs WHERE user_id = ? ORDER BY logged_at DESC LIMIT 1`, userID,
	).Scan(&w.ID, &w.UserID, &w.Weight, &w.BMI, &w.LoggedAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func GetPreviousWeightLog(userID int64) (*WeightLog, error) {
	w := &WeightLog{}
	err := db.QueryRow(
		`SELECT id, user_id, weight, bmi, logged_at FROM weight_logs WHERE user_id = ? ORDER BY logged_at DESC LIMIT 1 OFFSET 1`, userID,
	).Scan(&w.ID, &w.UserID, &w.Weight, &w.BMI, &w.LoggedAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func GetSession(userID int64) (*UserSession, error) {
	s := &UserSession{}
	err := db.QueryRow(
		`SELECT user_id, state, workout_id, updated_at FROM user_sessions WHERE user_id = ?`, userID,
	).Scan(&s.UserID, &s.State, &s.WorkoutID, &s.UpdatedAt)
	if err != nil {
		_, err2 := db.Exec(`INSERT OR IGNORE INTO user_sessions (user_id, state) VALUES (?, 'idle')`, userID)
		if err2 != nil {
			return nil, err2
		}
		s.UserID = userID
		s.State = "idle"
		s.WorkoutID = 0
		s.UpdatedAt = time.Now()
		return s, nil
	}
	return s, nil
}

func UpdateSession(userID int64, state string, workoutID int64) error {
	_, err := db.Exec(
		`INSERT INTO user_sessions (user_id, state, workout_id, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(user_id) DO UPDATE SET state = excluded.state, workout_id = excluded.workout_id, updated_at = CURRENT_TIMESTAMP`,
		userID, state, workoutID,
	)
	return err
}

func GetUserStats(userID int64) (*Stats, error) {
	user, err := GetUser(userID)
	if err != nil {
		return nil, err
	}

	bmi := CalculateBMI(user.Weight, user.Height)
	remaining := user.Weight - user.TargetWeight

	var lastWeightChange float64
	prevLog, err := GetPreviousWeightLog(userID)
	if err == nil {
		lastWeightChange = user.Weight - prevLog.Weight
	}

	weeklyLogs, err := GetWeeklyWorkoutLogs(userID)
	if err != nil {
		return nil, err
	}

	var totalScore float64
	var totalCalories int
	for _, l := range weeklyLogs {
		totalScore += l.Score
		totalCalories += l.Calories
	}

	var avgScore float64
	if len(weeklyLogs) > 0 {
		avgScore = totalScore / float64(len(weeklyLogs))
	}

	return &Stats{
		CurrentWeight:    user.Weight,
		CurrentBMI:       bmi,
		TargetWeight:     user.TargetWeight,
		WeightRemaining:  remaining,
		LastWeightChange: lastWeightChange,
		WeeklySessions:   len(weeklyLogs),
		WeeklyAvgScore:   avgScore,
		WeeklyCalories:   totalCalories,
		Streak:           user.Streak,
	}, nil
}

func GetUserPreferences(userID int64) (*UserPreferences, error) {
	p := &UserPreferences{}
	err := db.QueryRow(
		`SELECT user_id, goal, experience_level, has_dumbbell, has_resistance_band, has_pullup_bar, onboarding_done
		 FROM user_preferences WHERE user_id = ?`, userID,
	).Scan(&p.UserID, &p.Goal, &p.ExperienceLevel, &p.HasDumbbell, &p.HasResistanceBand, &p.HasPullupBar, &p.OnboardingDone)
	if err != nil {
		p.UserID = userID
		p.Goal = "diet"
		p.ExperienceLevel = "beginner"
		p.HasDumbbell = true
		p.HasResistanceBand = true
		p.HasPullupBar = true
		p.OnboardingDone = false
		return p, nil
	}
	return p, nil
}

func CreateUserPreferences(p *UserPreferences) error {
	_, err := db.Exec(
		`INSERT OR REPLACE INTO user_preferences (user_id, goal, experience_level, has_dumbbell, has_resistance_band, has_pullup_bar, onboarding_done)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.UserID, p.Goal, p.ExperienceLevel, boolToInt(p.HasDumbbell), boolToInt(p.HasResistanceBand), boolToInt(p.HasPullupBar), boolToInt(p.OnboardingDone),
	)
	return err
}

func UpdateUserPreferenceField(userID int64, field string, value interface{}) error {
	switch field {
	case "goal", "experience_level":
		_, err := db.Exec(
			`INSERT INTO user_preferences (user_id, `+field+`, onboarding_done) VALUES (?, ?, 0)
			 ON CONFLICT(user_id) DO UPDATE SET `+field+` = excluded.`+field,
			userID, value,
		)
		return err
	case "has_dumbbell", "has_resistance_band", "has_pullup_bar":
		_, err := db.Exec(
			`INSERT INTO user_preferences (user_id, `+field+`, onboarding_done) VALUES (?, ?, 0)
			 ON CONFLICT(user_id) DO UPDATE SET `+field+` = excluded.`+field,
			userID, value,
		)
		return err
	case "onboarding_done":
		_, err := db.Exec(
			`INSERT INTO user_preferences (user_id, onboarding_done) VALUES (?, ?)
			 ON CONFLICT(user_id) DO UPDATE SET onboarding_done = excluded.onboarding_done`,
			userID, value,
		)
		return err
	}
	return fmt.Errorf("unknown preference field: %s", field)
}

func CompleteOnboarding(userID int64) error {
	_, err := db.Exec(
		`INSERT INTO user_preferences (user_id, onboarding_done) VALUES (?, 1)
		 ON CONFLICT(user_id) DO UPDATE SET onboarding_done = 1`,
		userID,
	)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func CalculateBMI(weight, heightCm float64) float64 {
	heightM := heightCm / 100
	return weight / (heightM * heightM)
}

func SaveChatMessage(userID int64, role, content string) error {
	_, err := db.Exec(
		`INSERT INTO chat_history (user_id, role, content) VALUES (?, ?, ?)`,
		userID, role, content,
	)
	return err
}

func GetChatHistory(userID int64, limit int) ([]ChatMessage, error) {
	rows, err := db.Query(
		`SELECT id, user_id, role, content, created_at
		 FROM chat_history WHERE user_id = ? ORDER BY id DESC LIMIT ?`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func ClearOldChatHistory(userID int64, keepLast int) error {
	_, err := db.Exec(
		`DELETE FROM chat_history WHERE user_id = ? AND id NOT IN (
			SELECT id FROM chat_history WHERE user_id = ? ORDER BY id DESC LIMIT ?
		)`, userID, userID, keepLast,
	)
	return err
}

func SaveMoodLog(userID int64, mood string, energy int) error {
	_, err := db.Exec(
		`INSERT INTO mood_logs (user_id, mood, energy, logged_at) VALUES (?, ?, ?, datetime('now'))`,
		userID, mood, energy,
	)
	return err
}

func GetRecentMoodLogs(userID int64, limit int) ([]MoodLog, error) {
	rows, err := db.Query(
		`SELECT id, user_id, mood, energy, logged_at FROM mood_logs WHERE user_id = ? ORDER BY logged_at DESC LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []MoodLog
	for rows.Next() {
		var l MoodLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Mood, &l.Energy, &l.LoggedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}