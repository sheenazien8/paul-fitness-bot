package main

import (
	"time"
)

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

type UserPreferences struct {
	UserID            int64
	Goal              string
	ExperienceLevel   string
	HasDumbbell       bool
	HasResistanceBand bool
	HasPullupBar      bool
	OnboardingDone    bool
}

type LLMWorkoutResponse struct {
	Warmup   []Exercise `json:"warmup"`
	Main     []Exercise `json:"main"`
	Cooldown []Exercise `json:"cooldown"`
}

type Workout struct {
	ID          int64
	UserID      int64
	DayOfWeek   int
	WorkoutType string
	Exercises   string
	CreatedAt   time.Time
}

type WorkoutLog struct {
	ID              int64
	UserID          int64
	WorkoutID       int64
	DurationMinutes int
	Calories        int
	Satisfaction    int
	Score           float64
	LoggedAt        time.Time
}

type WeightLog struct {
	ID       int64
	UserID   int64
	Weight   float64
	BMI      float64
	LoggedAt time.Time
}

type UserSession struct {
	UserID    int64
	State     string
	WorkoutID int64
	UpdatedAt time.Time
}

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

// ChatMessage stores a single message in conversation history.
type ChatMessage struct {
	ID        int64
	UserID    int64
	Role      string `json:"role"` // "system", "user", "assistant", "tool"
	Content   string
	CreatedAt time.Time
}

// ToolCall represents a function call requested by the LLM.
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult is the result of executing a tool call
type ToolResult struct {
	ToolName string `json:"tool_name"`
	Success  bool   `json:"success"`
	Data     string `json:"data"`
	Error    string `json:"error,omitempty"`
}