package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultLLMBaseURL   = "http://localhost:11434"
	defaultLLMModel    = "glm-5.1:cloud"
	llmTimeout         = 60 * time.Second
	llmMaxRetries      = 2
	chatHistoryLimit   = 20
	rateLimitCooldown  = 3 * time.Second
)

type llmChatRequest struct {
	Model    string        `json:"model"`
	Messages []llmMsg   `json:"messages"`
	Stream   bool          `json:"stream"`
	Format   string        `json:"format,omitempty"`
	Tools    []llmTool  `json:"tools,omitempty"`
}

type llmMsg struct {
	Role      string        `json:"role"`
	Content   string        `json:"content"`
	ToolCalls []llmToolCall `json:"tool_calls,omitempty"`
}

type llmToolCall struct {
	Function llmFunctionCall `json:"function"`
}

type llmFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type llmTool struct {
	Type     string           `json:"type"`
	Function llmToolDef    `json:"function"`
}

type llmToolDef struct {
	Name        string             `json:"name"`
	Description  string             `json:"description"`
	Parameters  llmToolParams   `json:"parameters"`
}

type llmToolParams struct {
	Type       string                       `json:"type"`
	Required   []string                     `json:"required,omitempty"`
	Properties map[string]llmToolProp     `json:"properties"`
}

type llmToolProp struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type llmChatResponse struct {
	Model     string           `json:"model"`
	Message   llmMsg        `json:"message"`
	Done      bool             `json:"done"`
	Error     string           `json:"error,omitempty"`
}

type llmChatAPIResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role      string  `json:"role"`
			Content  string  `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func getToolDefinitions() []llmTool {
	return []llmTool{
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "generate_workout",
				Description: "Generate a workout routine for today based on the user's profile, preferences, and history. Returns structured exercise data (warmup, main, cooldown).",
				Parameters: llmToolParams{
					Type:     "object",
					Required: []string{},
					Properties: map[string]llmToolProp{
						"day_type": {
							Type:        "string",
							Description: "Override workout type (push/pull/legs/full_body). If not specified, uses the user's schedule for today.",
							Enum:        []string{"push", "pull", "legs", "full_body"},
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "log_weight",
				Description: "Log the user's body weight. Updates their current weight and BMI in the profile.",
				Parameters: llmToolParams{
					Type:     "object",
					Required: []string{"weight"},
					Properties: map[string]llmToolProp{
						"weight": {
							Type:        "number",
							Description: "Weight in kg",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "log_workout_done",
				Description: "Log a completed workout session with duration, calories burned, and satisfaction rating.",
				Parameters: llmToolParams{
					Type:     "object",
					Required: []string{"duration_minutes", "calories", "satisfaction"},
					Properties: map[string]llmToolProp{
						"duration_minutes": {
							Type:        "integer",
							Description: "Workout duration in minutes",
						},
						"calories": {
							Type:        "integer",
							Description: "Estimated calories burned",
						},
						"satisfaction": {
							Type:        "integer",
							Description: "Satisfaction rating 1-10",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "get_stats",
				Description: "Get the user's fitness stats including current weight, BMI, target progress, weekly workout stats, and streak.",
				Parameters: llmToolParams{
					Type:       "object",
					Required:   []string{},
					Properties: map[string]llmToolProp{},
				},
			},
		},
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "get_history",
				Description: "Get the user's recent workout history with scores and details.",
				Parameters: llmToolParams{
					Type:     "object",
					Required: []string{},
					Properties: map[string]llmToolProp{
						"limit": {
							Type:        "integer",
							Description: "Number of recent workouts to return (default 10)",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "get_user_profile",
				Description: "Get the user's full profile including personal data, preferences, equipment, and onboarding status.",
				Parameters: llmToolParams{
					Type:       "object",
					Required:   []string{},
					Properties: map[string]llmToolProp{},
				},
			},
		},
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "update_profile",
				Description: "Update the user's profile fields like weight, height, target weight. Also used during onboarding to set up initial data.",
				Parameters: llmToolParams{
					Type:     "object",
					Required: []string{},
					Properties: map[string]llmToolProp{
						"weight": {
							Type:        "number",
							Description: "Current weight in kg",
						},
						"height": {
							Type:        "number",
							Description: "Height in cm",
						},
						"target_weight": {
							Type:        "number",
							Description: "Target weight in kg",
						},
						"goal": {
							Type:        "string",
							Description: "Fitness goal",
							Enum:        []string{"diet", "muscle", "fitness", "maintenance"},
						},
						"experience_level": {
							Type:        "string",
							Description: "Experience level",
							Enum:        []string{"beginner", "intermediate", "advanced"},
						},
						"has_dumbbell": {
							Type:        "boolean",
							Description: "Has dumbbell equipment",
						},
						"has_resistance_band": {
							Type:        "boolean",
							Description: "Has resistance band equipment",
						},
						"has_pullup_bar": {
							Type:        "boolean",
							Description: "Has pull-up bar equipment",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "update_settings",
				Description: "Update the user's workout schedule and notification time.",
				Parameters: llmToolParams{
					Type:     "object",
					Required: []string{},
					Properties: map[string]llmToolProp{
						"workout_days": {
							Type:        "string",
							Description: "Comma-separated day numbers (1=Senin, 2=Selasa, 3=Rabu, 4=Kamis, 5=Jumat, 6=Sabtu, 7=Minggu)",
						},
						"notification_hour": {
							Type:        "integer",
							Description: "Hour for daily notification (0-23, WIB timezone)",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "complete_onboarding",
				Description: "Mark the user's onboarding as complete after all profile data has been collected.",
				Parameters: llmToolParams{
					Type:       "object",
					Required:   []string{},
					Properties: map[string]llmToolProp{},
				},
			},
		},
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "get_exercise_tips",
				Description: "Get form tips and how-to for a specific exercise from the workout pool. Use when user asks about proper form, technique, or alternatives for an exercise.",
				Parameters: llmToolParams{
					Type:       "object",
					Required:   []string{"exercise_name"},
					Properties: map[string]llmToolProp{
						"exercise_name": {
							Type:        "string",
							Description: "Name of the exercise to get tips for",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llmToolDef{
				Name:        "log_mood",
				Description: "Log the user's mood/energy level before or after workout. Helps track wellness and adjust workout intensity.",
				Parameters: llmToolParams{
					Type:       "object",
					Required:   []string{"mood", "energy"},
					Properties: map[string]llmToolProp{
						"mood": {
							Type:        "string",
							Description: "User's mood: great, good, okay, low, bad",
							Enum:        []string{"great", "good", "okay", "low", "bad"},
						},
						"energy": {
							Type:        "integer",
							Description: "Energy level from 1-10",
						},
					},
				},
			},
		},
	}
}

func buildSystemPrompt(userID int64) string {
	basePrompt := `Kamu adalah Paul, personal trainer AI yang friendly dan supportive. Bahasa utama Indonesia, casual pakai "kamu".

Kamu punya akses ke data user dan bisa melakukan aksi melalui tools. Gunakan tools ketika user meminta sesuatu yang butuh data/action, bukan saat hanya ngobrol.

Penting:
- Kalau user bilang "aku mau latihan" atau "hari ini latihan apa" → call generate_workout
- Kalau user bilang berat badan (contoh: "71.5", "aku 72kg", "berat aku 70") → call log_weight
- Kalau user bilang latihan selesai atau lapor durasi → call log_workout_done
- Kalau user tanya progress/stats → call get_stats
- Kalau user tanya riwayat → call get_history
- Kalau user tanya cara/form exercise → call get_exercise_tips
- Kalau user bilang mood/energi (contoh: "lagi semangat", "aku capek", "mood aku bagus") → call log_mood
- Jangan kasih nomor list kecuali untuk exercise list
- Exercise list format WAJIB konsisten (ini data, bukan kalimat):

🏋️ {Nama Hari}, {Tanggal} — {Tipe Latihan}

🔥 Pemanasan (5-7 menit)
• {Nama Exercise} = {Cara bahasa Indonesia} = {Reps}
...

💪 Latihan Utama
• {Nama Exercise} = {Cara bahasa Indonesia} = {Reps}
...

❄️ Cooling Down (5-10 menit)
• {Nama Exercise} = {Cara bahasa Indonesia} = {Reps}
...

- Pembuka & penutup boleh natural/beda tiap hari (motivasi, komentar score kemarin, dll)
- Exercise name dalam bahasa Inggris (standar gym)
- Cara/how_to dalam bahasa Indonesia
- Jangan halusinasi exercise — hanya gunakan data dari tool generate_workout
- Motivasi tapi gak cringe
- Jawab singkat, padat, gak bertele-tele
- Kalau user baru dan belum onboarding, bantu mereka setup profil. Tanya SATU PER SATU dengan urutan: (1) goal, (2) experience level, (3) alat yang dimiliki, (4) berat badan, (5) tinggi badan, (6) target berat, (7) hari latihan, (8) jam notifikasi. Setelah semua terisi, call complete_onboarding. JANGAN tanya semua sekaligus.
- Untuk hari latihan, gunakan angka: 1=Senin, 2=Selasa, 3=Rabu, 4=Kamis, 5=Jumat, 6=Sabtu, 7=Minggu`

	user, err := GetUser(userID)
	if err == nil && user != nil {
		bmi := CalculateBMI(user.Weight, user.Height)
		basePrompt += fmt.Sprintf("\n\n--- Data User ---\n")
		basePrompt += fmt.Sprintf("Nama: %s\n", user.FirstName)
		basePrompt += fmt.Sprintf("Berat: %.1f kg\n", user.Weight)
		basePrompt += fmt.Sprintf("Tinggi: %.0f cm\n", user.Height)
		basePrompt += fmt.Sprintf("BMI: %.1f\n", bmi)
		basePrompt += fmt.Sprintf("Target berat: %.1f kg\n", user.TargetWeight)
		basePrompt += fmt.Sprintf("Hari latihan: %s\n", user.WorkoutDays)
		basePrompt += fmt.Sprintf("Jam notifikasi: %02d:00 WIB\n", user.NotificationHour)
		basePrompt += fmt.Sprintf("Streak: %d hari\n", user.Streak)
	}

	prefs, err := GetUserPreferences(userID)
	if err == nil && prefs != nil {
		goalLabels := map[string]string{"diet": "Diet/Turun Berat", "muscle": "Bangun Otot", "fitness": "Fitness Umum", "maintenance": "Maintenance"}
		levelLabels := map[string]string{"beginner": "Pemula", "intermediate": "Menengah", "advanced": "Lanjutan"}

		goalLabel := goalLabels[prefs.Goal]
		if goalLabel == "" {
			goalLabel = prefs.Goal
		}
		levelLabel := levelLabels[prefs.ExperienceLevel]
		if levelLabel == "" {
			levelLabel = prefs.ExperienceLevel
		}

		basePrompt += fmt.Sprintf("\nTujuan: %s\n", goalLabel)
		basePrompt += fmt.Sprintf("Level: %s\n", levelLabel)
		basePrompt += fmt.Sprintf("Dumbbell: %v\n", prefs.HasDumbbell)
		basePrompt += fmt.Sprintf("Resistance Band: %v\n", prefs.HasResistanceBand)
		basePrompt += fmt.Sprintf("Pull-up Bar: %v\n", prefs.HasPullupBar)
		basePrompt += fmt.Sprintf("Onboarding selesai: %v\n", prefs.OnboardingDone)
	}

	return basePrompt
}

func executeTool(userID int64, call ToolCall) ToolResult {
	slog.Info("executing tool", "user_id", userID, "tool", call.Name, "args", call.Arguments)

	switch call.Name {
	case "generate_workout":
		return toolGenerateWorkout(userID, call.Arguments)
	case "log_weight":
		return toolLogWeight(userID, call.Arguments)
	case "log_workout_done":
		return toolLogWorkoutDone(userID, call.Arguments)
	case "get_stats":
		return toolGetStats(userID)
	case "get_history":
		return toolGetHistory(userID, call.Arguments)
	case "get_user_profile":
		return toolGetUserProfile(userID)
	case "update_profile":
		return toolUpdateProfile(userID, call.Arguments)
	case "update_settings":
		return toolUpdateSettings(userID, call.Arguments)
	case "complete_onboarding":
		return toolCompleteOnboarding(userID)
	case "get_exercise_tips":
		return toolGetExerciseTips(call.Arguments)
	case "log_mood":
		return toolLogMood(userID, call.Arguments)
	default:
		return ToolResult{ToolName: call.Name, Success: false, Error: fmt.Sprintf("unknown tool: %s", call.Name)}
	}
}

func toolGenerateWorkout(userID int64, args map[string]interface{}) ToolResult {
	now := time.Now()
	dayOfWeek := int(now.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7
	}

	dayType := GetWorkoutType(dayOfWeek)
	if dt, ok := args["day_type"].(string); ok && dt != "" {
		dayType = dt
	}

	if dayType == "" {
		return ToolResult{
			ToolName: "generate_workout",
			Success:  false,
			Error:    "Hari ini bukan hari latihan. Tidak ada workout yang tersedia.",
		}
	}

	existingWorkout, _ := GetTodaysWorkout(userID, dayOfWeek)
	var exercises []Exercise
	var warmUp []Exercise
	var coolDown []Exercise
	var workoutID int64

	if existingWorkout != nil {
		workoutID = existingWorkout.ID
		if err := json.Unmarshal([]byte(existingWorkout.Exercises), &exercises); err != nil {
			return ToolResult{ToolName: "generate_workout", Success: false, Error: "Gagal membaca data workout yang tersimpan."}
		}
		warmUp = GetWarmUp(dayType)
		coolDown = GetCoolDown(dayType)
	} else {
		llmWorkout := tryLLMWorkout(userID, dayType)
		if llmWorkout != nil {
			exercises = llmWorkout.Main
			warmUp = llmWorkout.Warmup
			cooldown := llmWorkout.Cooldown
			coolDown = cooldown
		} else {
			exercises = GetWorkoutForDay(dayOfWeek)
			warmUp = GetWarmUp(dayType)
			coolDown = GetCoolDown(dayType)
		}

		if len(exercises) == 0 {
			return ToolResult{ToolName: "generate_workout", Success: false, Error: "Tidak ada latihan yang tersedia untuk hari ini."}
		}

		var err error
		workoutID, err = SaveWorkout(userID, dayOfWeek, dayType, exercises)
		if err != nil {
			return ToolResult{ToolName: "generate_workout", Success: false, Error: "Gagal menyimpan workout."}
		}
	}

	typeName := WorkoutTypeNames[dayType]
	dayName := DayNames[dayOfWeek]
	dateStr := now.Format("2 January 2006")

	result := fmt.Sprintf("workout_id:%d|day_name:%s|date:%s|type:%s|type_name:%s\n",
		workoutID, dayName, dateStr, dayType, typeName)

	if len(warmUp) > 0 {
		result += "WARMUP_START\n"
		for _, ex := range warmUp {
			result += fmt.Sprintf("%s|%s|%s|%s\n", ex.Name, ex.HowTo, ex.Reps, ex.Muscle)
		}
		result += "WARMUP_END\n"
	}

	result += "MAIN_START\n"
	for _, ex := range exercises {
		result += fmt.Sprintf("%s|%s|%s|%s\n", ex.Name, ex.HowTo, ex.Reps, ex.Muscle)
	}
	result += "MAIN_END\n"

	if len(coolDown) > 0 {
		result += "COOLDOWN_START\n"
		for _, ex := range coolDown {
			result += fmt.Sprintf("%s|%s|%s|%s\n", ex.Name, ex.HowTo, ex.Reps, ex.Muscle)
		}
		result += "COOLDOWN_END\n"
	}

	return ToolResult{ToolName: "generate_workout", Success: true, Data: result}
}

func toolLogWeight(userID int64, args map[string]interface{}) ToolResult {
	weight, ok := args["weight"].(float64)
	if !ok {
		if intVal, ok := args["weight"].(int); ok {
			weight = float64(intVal)
		} else {
			return ToolResult{ToolName: "log_weight", Success: false, Error: "Berat badan tidak valid."}
		}
	}

	if weight < 20 || weight > 300 {
		return ToolResult{ToolName: "log_weight", Success: false, Error: "Berat badan tidak valid (harus 20-300 kg)."}
	}

	user, err := GetUser(userID)
	if err != nil {
		return ToolResult{ToolName: "log_weight", Success: false, Error: "User tidak ditemukan."}
	}

	bmi := CalculateBMI(weight, user.Height)
	if err := SaveWeightLog(userID, weight, bmi); err != nil {
		return ToolResult{ToolName: "log_weight", Success: false, Error: "Gagal menyimpan berat badan."}
	}

	weightChange := weight - user.Weight
	changeEmoji := "➡️"
	if weightChange < -0.5 {
		changeEmoji = "🎉"
	} else if weightChange < 0 {
		changeEmoji = "📉"
	} else if weightChange > 0.5 {
		changeEmoji = "📈"
	}

	remaining := weight - user.TargetWeight

	return ToolResult{
		ToolName: "log_weight",
		Success:  true,
		Data: fmt.Sprintf("Berat: %.1f kg | BMI: %.1f (%s %s) | Perubahan: %+.1f kg %s | Sisa ke target: %.1f kg",
			weight, bmi, GetBMIStatus(bmi), GetBMIEmoji(bmi), weightChange, changeEmoji, remaining),
	}
}

func toolLogWorkoutDone(userID int64, args map[string]interface{}) ToolResult {
	duration, _ := args["duration_minutes"].(float64)
	calories, _ := args["calories"].(float64)
	satisfaction, _ := args["satisfaction"].(float64)

	if duration <= 0 || calories <= 0 || satisfaction <= 0 {
		return ToolResult{ToolName: "log_workout_done", Success: false, Error: "Data tidak lengkap. Butuh duration_minutes, calories, dan satisfaction (1-10)."}
	}

	satInt := int(satisfaction)
	if satInt < 1 {
		satInt = 1
	}
	if satInt > 10 {
		satInt = 10
	}

	now := time.Now()
	dayOfWeek := int(now.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7
	}

	workout, err := GetTodaysWorkout(userID, dayOfWeek)
	if err != nil || workout == nil {
		workout, err = GetLatestWorkout(userID)
	}

	if workout == nil {
		return ToolResult{ToolName: "log_workout_done", Success: false, Error: "Tidak ada workout yang bisa dicatat. Gunakan generate_workout dulu."}
	}

	score := CalculateWorkoutScore(int(duration), int(calories), satInt)

	log := &WorkoutLog{
		UserID:           userID,
		WorkoutID:        workout.ID,
		DurationMinutes:  int(duration),
		Calories:         int(calories),
		Satisfaction:     satInt,
		Score:            score,
	}

	if err := SaveWorkoutLog(log); err != nil {
		return ToolResult{ToolName: "log_workout_done", Success: false, Error: "Gagal menyimpan log workout."}
	}

	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	user, _ := GetUser(userID)

	var newStreak int
	if user != nil {
		if user.LastWorkoutDate == yesterday {
			newStreak = user.Streak + 1
		} else if user.LastWorkoutDate == today {
			newStreak = user.Streak
		} else {
			newStreak = 1
		}
		UpdateUserStreak(userID, newStreak, today)
	}

	return ToolResult{
		ToolName: "log_workout_done",
		Success:  true,
		Data: fmt.Sprintf("Workout dicatat! Durasi: %d menit | Kalori: %d | Puas: %d/10 | Skor: %.1f %s | Streak: %d hari",
			int(duration), int(calories), satInt, score, GetScoreEmoji(score), newStreak),
	}
}

func toolGetStats(userID int64) ToolResult {
	stats, err := GetUserStats(userID)
	if err != nil {
		return ToolResult{ToolName: "get_stats", Success: false, Error: "Gagal mengambil statistik."}
	}

	bmiStatus := GetBMIStatus(stats.CurrentBMI)
	bmiEmoji := GetBMIEmoji(stats.CurrentBMI)

	weightChangeStr := "-"
	weightChangeEmoji := ""
	if stats.LastWeightChange != 0 {
		weightChangeStr = fmt.Sprintf("%+.1f kg", stats.LastWeightChange)
		weightChangeEmoji = GetWeightChangeEmoji(stats.LastWeightChange)
	}

	totalToLose := stats.CurrentWeight - stats.TargetWeight
	progressPct := 0.0
	if totalToLose > 0 {
		progressPct = (1 - stats.WeightRemaining/totalToLose) * 100
		if progressPct < 0 {
			progressPct = 0
		}
		if progressPct > 100 {
			progressPct = 100
		}
	}

	return ToolResult{
		ToolName: "get_stats",
		Success:  true,
		Data: fmt.Sprintf("Berat: %.1f kg | Target: %.1f kg | Sisa: %.1f kg | Perubahan terakhir: %s %s | BMI: %.1f %s %s | Progress: %.0f%% | Minggu ini: %d sesi, skor rata-rata %.1f, %d kalori | Streak: %d hari",
			stats.CurrentWeight, stats.TargetWeight, stats.WeightRemaining,
			weightChangeStr, weightChangeEmoji,
			stats.CurrentBMI, bmiStatus, bmiEmoji,
			progressPct,
			stats.WeeklySessions, stats.WeeklyAvgScore, stats.WeeklyCalories,
			stats.Streak),
	}
}

func toolGetHistory(userID int64, args map[string]interface{}) ToolResult {
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	logs, err := GetRecentWorkoutLogs(userID, limit)
	if err != nil {
		return ToolResult{ToolName: "get_history", Success: false, Error: "Gagal mengambil riwayat."}
	}

	if len(logs) == 0 {
		return ToolResult{ToolName: "get_history", Success: true, Data: "Belum ada riwayat latihan."}
	}

	var parts []string
	for _, l := range logs {
		parts = append(parts, fmt.Sprintf("%s: %d menit, %d kal, puas %d/10, skor %.1f %s",
			l.LoggedAt.Format("02 Jan 15:04"), l.DurationMinutes, l.Calories, l.Satisfaction, l.Score, GetScoreEmoji(l.Score)))
	}

	return ToolResult{
		ToolName: "get_history",
		Success:  true,
		Data:     strings.Join(parts, "\n"),
	}
}

func toolGetUserProfile(userID int64) ToolResult {
	user, err := GetUser(userID)
	if err != nil {
		return ToolResult{ToolName: "get_user_profile", Success: false, Error: "User tidak ditemukan."}
	}

	prefs, _ := GetUserPreferences(userID)
	bmi := CalculateBMI(user.Weight, user.Height)

	goalLabels := map[string]string{"diet": "Diet/Turun Berat", "muscle": "Bangun Otot", "fitness": "Fitness Umum", "maintenance": "Maintenance"}
	levelLabels := map[string]string{"beginner": "Pemula", "intermediate": "Menengah", "advanced": "Lanjutan"}

	goalLabel := goalLabels[prefs.Goal]
	if goalLabel == "" {
		goalLabel = prefs.Goal
	}
	levelLabel := levelLabels[prefs.ExperienceLevel]
	if levelLabel == "" {
		levelLabel = prefs.ExperienceLevel
	}

	var equipItems []string
	if prefs.HasDumbbell {
		equipItems = append(equipItems, "Dumbbell")
	}
	if prefs.HasResistanceBand {
		equipItems = append(equipItems, "Resistance Band")
	}
	if prefs.HasPullupBar {
		equipItems = append(equipItems, "Pull-Up Bar")
	}
	if len(equipItems) == 0 {
		equipItems = append(equipItems, "Bodyweight only")
	}

	return ToolResult{
		ToolName: "get_user_profile",
		Success:  true,
		Data: fmt.Sprintf("Nama: %s | Berat: %.1f kg | Tinggi: %.0f cm | BMI: %.1f (%s %s) | Target: %.1f kg | Tujuan: %s | Level: %s | Alat: %s | Hari latihan: %s | Notifikasi: %02d:00 | Onboarding: %v",
			user.FirstName, user.Weight, user.Height, bmi, GetBMIStatus(bmi), GetBMIEmoji(bmi),
			user.TargetWeight, goalLabel, levelLabel, strings.Join(equipItems, ", "),
			user.WorkoutDays, user.NotificationHour, prefs.OnboardingDone),
	}
}

func toolUpdateProfile(userID int64, args map[string]interface{}) ToolResult {
	user, err := GetUser(userID)
	if err != nil {
		return ToolResult{ToolName: "update_profile", Success: false, Error: "User tidak ditemukan."}
	}

	weight := user.Weight
	height := user.Height
	targetWeight := user.TargetWeight

	if w, ok := args["weight"].(float64); ok && w > 0 {
		weight = w
	} else if w, ok := args["weight"].(int); ok && w > 0 {
		weight = float64(w)
	}
	if h, ok := args["height"].(float64); ok && h > 0 {
		height = h
	} else if h, ok := args["height"].(int); ok && h > 0 {
		height = float64(h)
	}
	if t, ok := args["target_weight"].(float64); ok && t > 0 {
		targetWeight = t
	} else if t, ok := args["target_weight"].(int); ok && t > 0 {
		targetWeight = float64(t)
	}

	if err := UpdateUserProfile(userID, weight, height, targetWeight); err != nil {
		return ToolResult{ToolName: "update_profile", Success: false, Error: "Gagal update profil."}
	}

	if goal, ok := args["goal"].(string); ok {
		UpdateUserPreferenceField(userID, "goal", goal)
	}
	if level, ok := args["experience_level"].(string); ok {
		UpdateUserPreferenceField(userID, "experience_level", level)
	}
	if v, ok := args["has_dumbbell"].(bool); ok {
		UpdateUserPreferenceField(userID, "has_dumbbell", boolToInt(v))
	}
	if v, ok := args["has_resistance_band"].(bool); ok {
		UpdateUserPreferenceField(userID, "has_resistance_band", boolToInt(v))
	}
	if v, ok := args["has_pullup_bar"].(bool); ok {
		UpdateUserPreferenceField(userID, "has_pullup_bar", boolToInt(v))
	}

	return ToolResult{
		ToolName: "update_profile",
		Success:  true,
		Data:     fmt.Sprintf("Profil diupdate. Berat: %.1f kg, Tinggi: %.0f cm, Target: %.1f kg", weight, height, targetWeight),
	}
}

func toolUpdateSettings(userID int64, args map[string]interface{}) ToolResult {
	user, err := GetUser(userID)
	if err != nil {
		return ToolResult{ToolName: "update_settings", Success: false, Error: "User tidak ditemukan."}
	}

	workoutDays := user.WorkoutDays
	notifHour := user.NotificationHour

	if wd, ok := args["workout_days"].(string); ok && wd != "" {
		workoutDays = wd
	}
	if nh, ok := args["notification_hour"].(float64); ok {
		notifHour = int(nh)
	} else if nh, ok := args["notification_hour"].(int); ok {
		notifHour = nh
	}

	if err := UpdateUserSettings(userID, workoutDays, notifHour); err != nil {
		return ToolResult{ToolName: "update_settings", Success: false, Error: "Gagal update pengaturan."}
	}

	_ = UpdateUserSchedule(userID)

	return ToolResult{
		ToolName: "update_settings",
		Success:  true,
		Data:     fmt.Sprintf("Pengaturan diupdate. Hari latihan: %s, Jam notifikasi: %02d:00", workoutDays, notifHour),
	}
}

func toolCompleteOnboarding(userID int64) ToolResult {
	if err := CompleteOnboarding(userID); err != nil {
		return ToolResult{ToolName: "complete_onboarding", Success: false, Error: "Gagal menyelesaikan onboarding."}
	}
	return ToolResult{ToolName: "complete_onboarding", Success: true, Data: "Onboarding selesai!"}
}

func toolGetExerciseTips(args map[string]interface{}) ToolResult {
	name, ok := args["exercise_name"].(string)
	if !ok || name == "" {
		return ToolResult{ToolName: "get_exercise_tips", Success: false, Error: "Nama exercise tidak valid."}
	}

	name = strings.TrimSpace(name)
	for _, pool := range WorkoutPool {
		for _, group := range pool {
			for _, ex := range group {
				if strings.EqualFold(ex.Name, name) {
					return ToolResult{
						ToolName: "get_exercise_tips",
						Success:  true,
						Data:     fmt.Sprintf("%s: %s | Reps: %s | Otot: %s", ex.Name, ex.HowTo, ex.Reps, ex.Muscle),
					}
				}
			}
		}
	}

	for _, tips := range []struct {
		name   string
		howTo  string
		muscle string
	}{
		{"Push-Up", "Tangan selebar bahu, turunkan badan sampai dada hampir menyentuh lantai, dorong kembali. Jangan arch back!", "Chest, Triceps, Shoulders"},
		{"Plank", "Tahan posisi push-up atas dengan siku di bawah bahu. Jaga body straight line dari kepala sampai kaki.", "Core, Shoulders"},
		{"Squat", "Kaki selebar bahu, turunkan pinggul seperti mau duduk. Lutut tidak boleh melewati ujung kaki. Punggung tetap tegak.", "Quads, Glutes, Hamstrings"},
		{"Lunges", "Langkah maju satu kaki, turunkan pinggul sampai kedua lutut 90 derajat. Jangan biarkan lutut depan melewati jari kaki.", "Quads, Glutes"},
		{"Burpees", "Dari berdiri, squat down, lompat ke posisi plank, push-up, lompat kembali ke squat, lalu jump up.", "Full Body, Cardio"},
	} {
		if strings.EqualFold(tips.name, name) {
			return ToolResult{
					ToolName: "get_exercise_tips",
					Success:  true,
					Data:     fmt.Sprintf("%s: %s | Otot: %s", tips.name, tips.howTo, tips.muscle),
				}
		}
	}

	return ToolResult{ToolName: "get_exercise_tips", Success: false, Error: fmt.Sprintf("Exercise '%s' tidak ditemukan di database.", name)}
}

func toolLogMood(userID int64, args map[string]interface{}) ToolResult {
	mood, ok := args["mood"].(string)
	if !ok || mood == "" {
		return ToolResult{ToolName: "log_mood", Success: false, Error: "Mood tidak valid. Pilih: great, good, okay, low, bad"}
	}

	energy := 5
	if e, ok := args["energy"].(float64); ok {
		energy = int(e)
	} else if e, ok := args["energy"].(int); ok {
		energy = e
	}
	if energy < 1 {
		energy = 1
	}
	if energy > 10 {
		energy = 10
	}

	if err := SaveMoodLog(userID, mood, energy); err != nil {
		return ToolResult{ToolName: "log_mood", Success: false, Error: "Gagal menyimpan mood."}
	}

	moodLabels := map[string]string{"great": "Luar biasa! 🤩", "good": "Bagus! 😊", "okay": "Biasa aja 😐", "low": "Lagi nggak semangat 😔", "bad": "Lagi down 😢"}
	label := moodLabels[mood]
	if label == "" {
		label = mood
	}

	return ToolResult{
		ToolName: "log_mood",
		Success:  true,
		Data:     fmt.Sprintf("Mood tercatat: %s | Energy: %d/10. %s", label, energy, moodRecommendation(mood, energy)),
	}
}

func moodRecommendation(mood string, energy int) string {
	if mood == "bad" || mood == "low" || energy <= 3 {
		return "Mungkin latihan ringan aja hari ini, atau istirahat total juga oke!"
	}
	if energy <= 5 {
		return "Coba latihan moderate aja, jangan push terlalu keras."
	}
	return "Sepertinya kamu siap buat latihan keras! Gas! 💪"
}

func tryLLMWorkout(userID int64, dayType string) *LLMWorkoutResponse {
	user, err := GetUser(userID)
	if err != nil {
		slog.Error("get user for LLM workout failed", "error", err)
		return nil
	}

	prefs, err := GetUserPreferences(userID)
	if err != nil {
		slog.Error("get preferences for LLM workout failed", "error", err)
		return nil
	}

	recentLogs, _ := GetRecentWorkoutLogs(userID, 3)

	llmWorkout, err := GenerateWorkoutWithLLM(user, prefs, dayType, recentLogs)
	if err != nil {
		slog.Warn("LLM workout generation failed, falling back to pool", "error", err)
		return nil
	}

	return llmWorkout
}

func GenerateWorkoutWithLLM(user *User, prefs *UserPreferences, dayType string, recentLogs []WorkoutLog) (*LLMWorkoutResponse, error) {
	prompt := buildWorkoutPrompt(user, prefs, dayType, recentLogs)

	result, err := callLLMGenerate(prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	var llmResp LLMWorkoutResponse
	if err := json.Unmarshal([]byte(result), &llmResp); err != nil {
		slog.Error("failed to parse LLM JSON response", "error", err, "raw", truncateString(result, 500))
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}

	if !validateLLMResponse(&llmResp) {
		slog.Error("LLM response validation failed",
			"warmup_count", len(llmResp.Warmup),
			"main_count", len(llmResp.Main),
			"cooldown_count", len(llmResp.Cooldown),
		)
		return nil, fmt.Errorf("LLM response validation failed")
	}

	return &llmResp, nil
}

func buildWorkoutPrompt(user *User, prefs *UserPreferences, dayType string, recentLogs []WorkoutLog) string {
	goalDesc := map[string]string{
		"diet":        "Diet/Turun Berat (fokus fat loss, calorie burn)",
		"muscle":      "Bangun Otot (fokus hypertrophy, strength)",
		"fitness":     "Fitness Umum (kesehatan overall, stamina)",
		"maintenance": "Maintenance (jaga berat & fitness saat ini)",
	}

	levelDesc := map[string]string{
		"beginner":     "Pemula (baru mulai latihan, perlu gerakan dasar)",
		"intermediate": "Menengah (sudah rutin latihan 3-6 bulan)",
		"advanced":     "Lanjutan (latihan rutin >1 tahun, familiar dengan teknik lanjutan)",
	}

	dayTypeDesc := map[string]string{
		"push":      "Push — Dada, Bahu, Tricep",
		"pull":      "Pull — Punggung, Bicep",
		"legs":      "Legs — Kaki, Glute",
		"full_body": "Full Body — Seluruh Tubuh",
	}

	bmi := CalculateBMI(user.Weight, user.Height)

	var equipmentItems []string
	if prefs.HasDumbbell {
		equipmentItems = append(equipmentItems, "Dumbbell")
	}
	if prefs.HasResistanceBand {
		equipmentItems = append(equipmentItems, "Resistance Band")
	}
	if prefs.HasPullupBar {
		equipmentItems = append(equipmentItems, "Pull-Up Bar")
	}
	if len(equipmentItems) == 0 {
		equipmentItems = append(equipmentItems, "Tidak ada (bodyweight only)")
	}
	equipmentList := strings.Join(equipmentItems, ", ")

	historyContext := ""
	if len(recentLogs) > 0 {
		var parts []string
		for i, l := range recentLogs {
			if i >= 3 {
				break
			}
			parts = append(parts, fmt.Sprintf("- Skor %.1f/100 (durasi %d menit, kalori %d, puas %d/10)",
				l.Score, l.DurationMinutes, l.Calories, l.Satisfaction))
		}
		historyContext = fmt.Sprintf("\n\nRiwayat latihan terakhir:\n%s\n", strings.Join(parts, "\n"))

		avgScore := 0.0
		for _, l := range recentLogs {
			if len(recentLogs) <= 3 {
				avgScore += l.Score
			}
		}
		if len(recentLogs) > 0 {
			avgScore /= float64(min(len(recentLogs), 3))
		}
		if avgScore > 0 {
			if avgScore < 40 {
				historyContext += "\nCatatan: Score rata-rata rendah, kurangi intensity dan pilih gerakan yang lebih mudah.\n"
			} else if avgScore > 75 {
				historyContext += "\nCatatan: Score rata-rata tinggi, boleh tingkatkan pelan intensity dan tambah beban/reps.\n"
			}
		}
	}

	goalText := goalDesc[prefs.Goal]
	if goalText == "" {
		goalText = goalDesc["diet"]
	}
	levelText := levelDesc[prefs.ExperienceLevel]
	if levelText == "" {
		levelText = levelDesc["beginner"]
	}
	dayText := dayTypeDesc[dayType]
	if dayText == "" {
		dayText = dayType
	}

	prompt := fmt.Sprintf(`Kamu adalah pelatih fitness profesional. Buatkan rutinitas latihan rumahan dengan format JSON.

Info user:
- Tujuan: %s
- Level: %s
- Alat: %s
- Berat: %.1f kg, Tinggi: %.0f cm, BMI: %.1f
- Target: %.1f kg

Hari ini: %s (%s)%s

Buatkan:
1. 4-5 latihan pemanasan (warm-up) yang sesuai
2. 6-8 latihan utama untuk %s
3. 4-5 stretching cooling down

Format JSON response (WAJIB JSON, jangan tambah teks lain):
{
  "warmup": [
    {"name": "Nama Exercise", "how_to": "Cara dalam bahasa Indonesia", "reps": "3x12", "muscle": "Target muscle"}
  ],
  "main": [
    {"name": "Nama Exercise", "how_to": "Cara dalam bahasa Indonesia", "reps": "3x12", "muscle": "Target muscle"}
  ],
  "cooldown": [
    {"name": "Nama Exercise", "how_to": "Cara dalam bahasa Indonesia", "reps": "2x20d", "muscle": "Target muscle"}
  ]
}

Penting:
- Exercise name dalam bahasa Inggris (standar gym)
- Cara/how_to dalam bahasa Indonesia
- Reps sesuai level user (beginner: lebih banyak rep ringan, advanced: lebih berat fewer rep)
- Sesuaikan intensity berdasarkan riwayat (score rendah = kurangi, score tinggi = tingkatkan pelan)
- Fokus fat loss kalau goal = diet (lebih banyak compound movement, shorter rest, higher rep)
- Hanya gunakan alat yang tersedia
- Jangan tambah teks apapun selain JSON`, goalText, levelText, equipmentList, user.Weight, user.Height, bmi, user.TargetWeight, dayType, dayText, historyContext, dayType)

	return prompt
}

func validateLLMResponse(resp *LLMWorkoutResponse) bool {
	if len(resp.Warmup) < 3 || len(resp.Warmup) > 6 {
		return false
	}
	if len(resp.Main) < 5 || len(resp.Main) > 10 {
		return false
	}
	if len(resp.Cooldown) < 3 || len(resp.Cooldown) > 6 {
		return false
	}

	for _, ex := range resp.Warmup {
		if ex.Name == "" || ex.Reps == "" {
			return false
		}
	}
	for _, ex := range resp.Main {
		if ex.Name == "" || ex.Reps == "" {
			return false
		}
	}
	for _, ex := range resp.Cooldown {
		if ex.Name == "" || ex.Reps == "" {
			return false
		}
	}

	return true
}

var userLastRequest = make(map[int64]time.Time)
var rateLimitMu sync.Mutex

func isRateLimited(userID int64) bool {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	last, exists := userLastRequest[userID]
	if exists && time.Since(last) < rateLimitCooldown {
		return true
	}
	userLastRequest[userID] = time.Now()
	return false
}

func ChatWithLLM(userID int64, userMessage string) string {
	if isRateLimited(userID) {
		return "Sabar ya, tunggu bentar lagi aku balas! ⏳"
	}

	systemPrompt := buildSystemPrompt(userID)

	history, err := GetChatHistory(userID, chatHistoryLimit)
	if err != nil {
		slog.Error("get chat history failed", "user_id", userID, "error", err)
		history = nil
	}

	messages := []llmMsg{
		{Role: "system", Content: systemPrompt},
	}

	for _, msg := range history {
		messages = append(messages, llmMsg{Role: msg.Role, Content: msg.Content})
	}

	messages = append(messages, llmMsg{Role: "user", Content: userMessage})

	SaveChatMessage(userID, "user", userMessage)

	startTime := time.Now()
	maxRounds := 3
	for round := 0; round < maxRounds; round++ {
		response, toolCalls, err := callLLMChatWithTools(messages)
		if err != nil {
			slog.Error("LLM chat failed", "round", round, "duration", time.Since(startTime), "error", err)
			fallbackMsg := "Maaf, aku lagi gangguan nih. Coba lagi ya! 🙏"
			SaveChatMessage(userID, "assistant", fallbackMsg)
			return fallbackMsg
		}

		if len(toolCalls) == 0 {
			assistantContent := response
			SaveChatMessage(userID, "assistant", assistantContent)
			ClearOldChatHistory(userID, chatHistoryLimit)
			return assistantContent
		}

		assistantMsg := llmMsg{
			Role:      "assistant",
			Content:   response,
			ToolCalls: make([]llmToolCall, len(toolCalls)),
		}
		for i, tc := range toolCalls {
			argsBytes, _ := json.Marshal(tc.Arguments)
			assistantMsg.ToolCalls[i] = llmToolCall{
				Function: llmFunctionCall{
					Name:      tc.Name,
					Arguments: string(argsBytes),
				},
			}
		}
		messages = append(messages, assistantMsg)

		for _, tc := range toolCalls {
			result := executeTool(userID, tc)

			resultJSON, err := json.Marshal(map[string]interface{}{
				"tool_name": result.ToolName,
				"success":   result.Success,
				"data":      result.Data,
				"error":     result.Error,
			})
			if err != nil {
				slog.Error("marshal tool result failed", "tool", tc.Name, "error", err)
				resultJSON = []byte(`{"tool_name":"` + tc.Name + `","success":false,"error":"internal error"}`)
			}

			messages = append(messages, llmMsg{
				Role:    "tool",
				Content: string(resultJSON),
			})

			slog.Info("tool executed", "tool", tc.Name, "success", result.Success, "duration", time.Since(startTime))
		}

		SaveChatMessage(userID, "assistant", response)

		// Continue loop — LLM will respond with the tool results incorporated
	}

	finalResponse, _, err := callLLMChatWithTools(messages)
	if err != nil {
		fallbackMsg := "Maaf, ada masalah teknis. Coba lagi nanti ya! 🙏"
		SaveChatMessage(userID, "assistant", fallbackMsg)
		return fallbackMsg
	}

	SaveChatMessage(userID, "assistant", finalResponse)
	ClearOldChatHistory(userID, chatHistoryLimit)
	return finalResponse
}

func getLLMBaseURL() string {
	if url := os.Getenv("LLM_BASE_URL"); url != "" {
		return strings.TrimRight(url, "/")
	}
	return defaultLLMBaseURL
}

func getLLMModel() string {
	if model := os.Getenv("LLM_MODEL"); model != "" {
		return model
	}
	return defaultLLMModel
}

func getLLMAPIKey() string {
	return os.Getenv("LLM_API_KEY")
}

func callLLMGenerate(prompt string) (string, error) {
	url := getLLMBaseURL() + "/v1/chat/completions"
	model := getLLMModel()
	apiKey := getLLMAPIKey()

	systemPrompt := "Kamu adalah pelatih fitness profesional. Selalu jawab dengan JSON valid sesuai permintaan. Jangan tambah teks lain di luar JSON."
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"stream": false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	reqStart := time.Now()
	slog.Info("calling LLM generate", "url", url, "model", model, "prompt_len", len(prompt))

	client := &http.Client{Timeout: llmTimeout}
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	slog.Info("LLM generate response", "status", resp.StatusCode, "duration", time.Since(reqStart))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp llmChatAPIResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if chatResp.Error.Message != "" {
		return "", fmt.Errorf("LLM error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func callLLMChatWithTools(messages []llmMsg) (string, []ToolCall, error) {
	var lastErr error
	for attempt := 0; attempt < llmMaxRetries; attempt++ {
		if attempt > 0 {
			slog.Info("retrying LLM call", "attempt", attempt+1)
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		content, toolCalls, err := doLLMChatRequest(messages)
		if err == nil {
			return content, toolCalls, nil
		}
		lastErr = err
		slog.Warn("LLM call failed", "attempt", attempt+1, "error", err)
	}
	return "", nil, fmt.Errorf("LLM unavailable after %d retries: %w", llmMaxRetries, lastErr)
}

func doLLMChatRequest(messages []llmMsg) (string, []ToolCall, error) {
	url := getLLMBaseURL() + "/v1/chat/completions"
	model := getLLMModel()
	apiKey := getLLMAPIKey()

	reqBody := llmChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Tools:    getToolDefinitions(),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("marshal request: %w", err)
	}

	reqStart := time.Now()
	slog.Info("calling LLM chat", "url", url, "model", model, "messages", len(messages))

	client := &http.Client{Timeout: llmTimeout}
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read response body: %w", err)
	}

	slog.Info("LLM response received", "status", resp.StatusCode, "duration", time.Since(reqStart))

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("LLM returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp llmChatAPIResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		slog.Error("failed to parse chat response", "error", err, "raw", truncateString(string(bodyBytes), 500))
		return "", nil, fmt.Errorf("decode response: %w", err)
	}

	if chatResp.Error.Message != "" {
		return "", nil, fmt.Errorf("LLM error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", nil, fmt.Errorf("no response choices")
	}

	msg := chatResp.Choices[0].Message
	var toolCalls []ToolCall
	if len(msg.ToolCalls) > 0 {
		for _, tc := range msg.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				slog.Warn("failed to parse tool call arguments", "tool", tc.Function.Name, "raw_args", tc.Function.Arguments, "error", err)
				args = make(map[string]interface{})
			}
			toolCalls = append(toolCalls, ToolCall{
				Name:      tc.Function.Name,
				Arguments: args,
			})
		}
	}

	return msg.Content, toolCalls, nil
}

func FormatWorkoutForDisplay(workoutID int64, dayName, dateStr, dayType, typeName string, warmUp, exercises, coolDown []Exercise) string {
	text := fmt.Sprintf("🏋️ <b>%s, %s — %s</b>\n\n", dayName, dateStr, typeName)

	if len(warmUp) > 0 {
		text += "🔥 <b>Pemanasan (5-7 menit):</b>\n\n"
		for _, ex := range warmUp {
			text += fmt.Sprintf("• <b>%s</b> = %s = %s\n", ex.Name, ex.HowTo, ex.Reps)
		}
		text += "\n"
	}

	text += "💪 <b>Latihan Utama:</b>\n\n"
	for i, ex := range exercises {
		text += fmt.Sprintf("• <b>%s</b> = %s = %s\n", ex.Name, ex.HowTo, ex.Reps)
		if i < len(exercises)-1 {
			text += "\n"
		}
	}

	if len(coolDown) > 0 {
		text += "\n❄️ <b>Cooling Down (5-10 menit):</b>\n\n"
		for _, ex := range coolDown {
			text += fmt.Sprintf("• <b>%s</b> = %s = %s\n", ex.Name, ex.HowTo, ex.Reps)
		}
	}

	return text
}

func ParseWorkoutToolResult(data string) (workoutID int64, dayName, dateStr, typeName string, warmUp, main, coolDown []Exercise, err error) {
	lines := strings.Split(data, "\n")
	if len(lines) == 0 {
		return 0, "", "", "", nil, nil, nil, fmt.Errorf("empty data")
	}

	header := lines[0]
	parts := strings.Split(header, "|")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "workout_id":
			fmt.Sscanf(kv[1], "%d", &workoutID)
		case "day_name":
			dayName = kv[1]
		case "date":
			dateStr = kv[1]
		case "type_name":
			typeName = kv[1]
		}
	}

	var currentSection string
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch line {
		case "WARMUP_START":
			currentSection = "warmup"
		case "WARMUP_END", "MAIN_END", "COOLDOWN_END":
			currentSection = ""
		case "MAIN_START":
			currentSection = "main"
		case "COOLDOWN_START":
			currentSection = "cooldown"
		default:
			exParts := strings.SplitN(line, "|", 4)
			if len(exParts) < 3 {
				continue
			}
			ex := Exercise{
				Name:  exParts[0],
				HowTo: exParts[1],
				Reps:  exParts[2],
			}
			if len(exParts) > 3 {
				ex.Muscle = exParts[3]
			}
			switch currentSection {
			case "warmup":
				warmUp = append(warmUp, ex)
			case "main":
				main = append(main, ex)
			case "cooldown":
				coolDown = append(coolDown, ex)
			}
		}
	}

	return
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

type llmGenerateResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}