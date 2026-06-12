package main

import (
	"time"
)

// User represents a bot user
type User struct {
	UserID           int64
	Username         string
	FirstName        string
	Weight           float64
	Height           float64
	TargetWeight     float64
	WorkoutDays      string
	NotificationHour int
	Streak           int
	LastWorkoutDate  string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// UserPreferences represents user onboarding choices
type UserPreferences struct {
	UserID            int64
	Goal              string // diet, muscle, fitness, maintenance
	ExperienceLevel   string // beginner, intermediate, advanced
	HasDumbbell       bool
	HasResistanceBand bool
	HasPullupBar      bool
	OnboardingDone    bool
}

// LLMWorkoutResponse represents the JSON response from Ollama
type LLMWorkoutResponse struct {
	Warmup   []Exercise `json:"warmup"`
	Main     []Exercise `json:"main"`
	Cooldown []Exercise `json:"cooldown"`
}

// Workout represents a scheduled workout
type Workout struct {
	ID          int64
	UserID      int64
	DayOfWeek   int
	WorkoutType string
	Exercises   string // JSON
	CreatedAt   time.Time
}

// WorkoutLog represents a completed workout entry
type WorkoutLog struct {
	ID              int64
	UserID          int64
	WorkoutID       int64
	DurationMinutes int
	Calories        int
	Satisfaction     int
	Score           float64
	LoggedAt        time.Time
}

// WeightLog represents a weight entry
type WeightLog struct {
	ID       int64
	UserID   int64
	Weight   float64
	BMI      float64
	LoggedAt time.Time
}

// UserSession tracks pending input state for a user
type UserSession struct {
	UserID    int64
	State     string // "idle", "awaiting_workout_log", "awaiting_weight", onboarding states
	WorkoutID int64
	UpdatedAt time.Time
}

// Stats represents aggregated user statistics
type Stats struct {
	CurrentWeight    float64
	CurrentBMI       float64
	TargetWeight     float64
	WeightRemaining  float64
	LastWeightChange float64
	WeeklySessions   int
	WeeklyAvgScore   float64
	WeeklyCalories   int
	Streak           int
}