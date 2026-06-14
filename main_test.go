package main

import (
	"strings"
	"testing"
)

func TestCalculateWorkoutScore(t *testing.T) {
	tests := []struct {
		name       string
		duration   int
		calories   int
		satisfaction int
		want       float64
	}{
		{"perfect workout", 60, 500, 10, 100.0},
		{"good workout", 45, 350, 8, 75.5},
		{"short workout", 15, 100, 5, 33.5},
		{"zero values", 0, 0, 0, 0.0},
		{"very long workout capped", 120, 800, 10, 100.0},
		{"high calories capped", 60, 1000, 10, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateWorkoutScore(tt.duration, tt.calories, tt.satisfaction)
			if got != tt.want {
				t.Errorf("CalculateWorkoutScore(%d, %d, %d) = %.1f, want %.1f", tt.duration, tt.calories, tt.satisfaction, got, tt.want)
			}
		})
	}
}

func TestGetScoreEmoji(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{95.0, "🔥"},
		{90.0, "🔥"},
		{80.0, "💪"},
		{75.0, "💪"},
		{65.0, "👍"},
		{60.0, "👍"},
		{50.0, "🙂"},
		{40.0, "🙂"},
		{30.0, "😅"},
		{0.0, "😅"},
	}
	for _, tt := range tests {
		got := GetScoreEmoji(tt.score)
		if got != tt.want {
			t.Errorf("GetScoreEmoji(%.1f) = %s, want %s", tt.score, got, tt.want)
		}
	}
}

func TestGetBMIStatus(t *testing.T) {
	tests := []struct {
		bmi  float64
		want string
	}{
		{16.0, "Kurus"},
		{18.4, "Kurus"},
		{18.5, "Normal"},
		{22.0, "Normal"},
		{24.9, "Normal"},
		{25.0, "Gemuk"},
		{28.0, "Gemuk"},
		{29.9, "Gemuk"},
		{30.0, "Obesitas"},
		{35.0, "Obesitas"},
	}
	for _, tt := range tests {
		got := GetBMIStatus(tt.bmi)
		if got != tt.want {
			t.Errorf("GetBMIStatus(%.1f) = %s, want %s", tt.bmi, got, tt.want)
		}
	}
}

func TestCalculateBMI(t *testing.T) {
	tests := []struct {
		weight   float64
		heightCm float64
		want     float64
	}{
		{70, 170, 24.22},
		{60, 160, 23.44},
		{100, 180, 30.86},
	}
	for _, tt := range tests {
		got := CalculateBMI(tt.weight, tt.heightCm)
		diff := got - tt.want
		if diff < -0.1 || diff > 0.1 {
			t.Errorf("CalculateBMI(%.0f, %.0f) = %.2f, want %.2f", tt.weight, tt.heightCm, got, tt.want)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		maxLen  int
		want    string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"long string", "hello world", 5, "hello..."},
		{"empty string", "", 5, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestGetWorkoutType(t *testing.T) {
	tests := []struct {
		day  int
		want string
	}{
		{1, "push"},
		{2, "legs"},
		{3, ""},
		{4, "pull"},
		{5, "full_body"},
		{6, ""},
		{7, ""},
	}
	for _, tt := range tests {
		got := GetWorkoutType(tt.day)
		if got != tt.want {
			t.Errorf("GetWorkoutType(%d) = %q, want %q", tt.day, got, tt.want)
		}
	}
}

func TestParseWorkoutDays(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []int
	}{
		{"single day", "1", []int{1}},
		{"multiple days", "1,3,5", []int{1, 3, 5}},
		{"with spaces", " 1 , 3 , 5 ", []int{1, 3, 5}},
		{"empty", "", nil},
		{"invalid", "abc", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseWorkoutDays(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseWorkoutDays(%q) = %v, want %v", tt.input, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseWorkoutDays(%q)[%d] = %d, want %d", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFormatWorkoutForDisplay(t *testing.T) {
	warmUp := []Exercise{
		{Name: "Arm Circles", HowTo: "Putar lengan", Reps: "2x20", Muscle: "Shoulders"},
	}
	exercises := []Exercise{
		{Name: "Bench Press", HowTo: "Dorong ke atas", Reps: "3x12", Muscle: "Chest"},
	}
	coolDown := []Exercise{
		{Name: "Chest Stretch", HowTo: "Regang dada", Reps: "2x30d", Muscle: "Chest"},
	}

	result := FormatWorkoutForDisplay(1, "Senin", "16 Juni 2025", "push", "Push (Dada, Bahu, Tricep)", warmUp, exercises, coolDown)

	if result == "" {
		t.Error("FormatWorkoutForDisplay returned empty string")
	}
	if !strings.Contains(result, "Bench Press") {
		t.Error("FormatWorkoutForDisplay missing exercise name")
	}
	if !strings.Contains(result, "1") {
		t.Error("FormatWorkoutForDisplay missing workout ID for button")
	}
}

func TestParseWorkoutToolResult(t *testing.T) {
	data := "workout_id:42|day_name:Senin|date:16 Juni 2025|type:push|type_name:Push\nWARMUP_START\nArm Circles|Putar lengan|2x20|Shoulders\nWARMUP_END\nMAIN_START\nBench Press|Dorong ke atas|3x12|Chest\nMAIN_END"

	workoutID, dayName, dateStr, typeName, warmUp, main, coolDown, err := ParseWorkoutToolResult(data)
	if err != nil {
		t.Fatalf("ParseWorkoutToolResult returned error: %v", err)
	}
	if workoutID != 42 {
		t.Errorf("workoutID = %d, want 42", workoutID)
	}
	if dayName != "Senin" {
		t.Errorf("dayName = %q, want %q", dayName, "Senin")
	}
	if len(warmUp) != 1 || warmUp[0].Name != "Arm Circles" {
		t.Errorf("warmUp = %v, want 1 exercise", warmUp)
	}
	if len(main) != 1 || main[0].Name != "Bench Press" {
		t.Errorf("main = %v, want 1 exercise", main)
	}
	if len(coolDown) != 0 {
		t.Errorf("coolDown = %v, want empty", coolDown)
	}
	_ = dateStr
	_ = typeName
}

func TestGetScoreDescription(t *testing.T) {
	tests := []struct {
		score float64
		want string
	}{
		{95.0, "Luar biasa! Kamu memberikan yang terbaik!"},
		{80.0, "Hebat! Latihan yang solid!"},
		{65.0, "Bagus! Terus pertahankan!"},
		{45.0, "Lumayan, semangat lagi ya!"},
		{20.0, "Yuk semangat! Setiap langih berarti!"},
	}
	for _, tt := range tests {
		got := GetScoreDescription(tt.score)
		if got != tt.want {
			t.Errorf("GetScoreDescription(%.1f) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

func TestGetBMIEmoji(t *testing.T) {
	tests := []struct {
		bmi  float64
		want string
	}{
		{17.0, "📉"},
		{22.0, "✅"},
		{27.0, "⚠️"},
		{32.0, "🔴"},
	}
	for _, tt := range tests {
		got := GetBMIEmoji(tt.bmi)
		if got != tt.want {
			t.Errorf("GetBMIEmoji(%.1f) = %s, want %s", tt.bmi, got, tt.want)
		}
	}
}

func TestGetWeightChangeEmoji(t *testing.T) {
	tests := []struct {
		change float64
		want   string
	}{
		{-1.0, "🎉"},
		{-0.3, "📉"},
		{0.0, "➡️"},
		{0.3, "📊"},
		{1.0, "📈"},
	}
	for _, tt := range tests {
		got := GetWeightChangeEmoji(tt.change)
		if got != tt.want {
			t.Errorf("GetWeightChangeEmoji(%.1f) = %s, want %s", tt.change, got, tt.want)
		}
	}
}

func TestGetWarmUp(t *testing.T) {
	tests := []struct {
		workoutType string
		wantLen     bool
	}{
		{"push", true},
		{"pull", true},
		{"legs", true},
		{"full_body", true},
		{"unknown", false},
	}
	for _, tt := range tests {
		got := GetWarmUp(tt.workoutType)
		if (len(got) > 0) != tt.wantLen {
			t.Errorf("GetWarmUp(%q) returned %d exercises, wantLen=%v", tt.workoutType, len(got), tt.wantLen)
		}
	}
}

func TestGetCoolDown(t *testing.T) {
	tests := []struct {
		workoutType string
		wantLen     bool
	}{
		{"push", true},
		{"pull", true},
		{"legs", true},
		{"full_body", true},
		{"unknown", false},
	}
	for _, tt := range tests {
		got := GetCoolDown(tt.workoutType)
		if (len(got) > 0) != tt.wantLen {
			t.Errorf("GetCoolDown(%q) returned %d exercises, wantLen=%v", tt.workoutType, len(got), tt.wantLen)
		}
	}
}

func TestGetWorkoutForDay(t *testing.T) {
	push := GetWorkoutForDay(1)
	if push == nil {
		t.Error("GetWorkoutForDay(1) returned nil for push day")
	}
	off := GetWorkoutForDay(3)
	if off != nil {
		t.Errorf("GetWorkoutForDay(3) = %v, want nil for off day", off)
	}
}

func TestIsRateLimited(t *testing.T) {
	userID := int64(99999)

	if isRateLimited(userID) {
		t.Error("first call should not be rate limited")
	}
	if !isRateLimited(userID) {
		t.Error("second call within 3s should be rate limited")
	}
}

func TestMoodRecommendation(t *testing.T) {
	tests := []struct {
		mood    string
		energy  int
		wantSub string
	}{
		{"bad", 2, "istirahat"},
		{"low", 3, "ringan"},
		{"okay", 5, "moderate"},
		{"good", 7, "siap"},
		{"great", 9, "keras"},
	}
	for _, tt := range tests {
		got := moodRecommendation(tt.mood, tt.energy)
		if !strings.Contains(got, tt.wantSub) {
			t.Errorf("moodRecommendation(%q, %d) = %q, want substring %q", tt.mood, tt.energy, got, tt.wantSub)
		}
	}
}

func TestFormatWorkoutForDisplayEmpty(t *testing.T) {
	result := FormatWorkoutForDisplay(1, "Senin", "16 Juni", "push", "Push", nil, nil, nil)
	if result == "" {
		t.Error("FormatWorkoutForDisplay with empty exercises should not be empty")
	}
}

func TestParseWorkoutToolResultMinimal(t *testing.T) {
	data := "workout_id:1|day_name:Senin|date:16 Juni|type:push|type_name:Push"
	_, _, _, _, _, _, _, err := ParseWorkoutToolResult(data)
	if err != nil {
		t.Fatalf("ParseWorkoutToolResult minimal should not error: %v", err)
	}
}

func TestParseWorkoutToolResultInvalidHeader(t *testing.T) {
	data := "garbage_data"
	workoutID, _, _, _, _, _, _, err := ParseWorkoutToolResult(data)
	if err != nil {
		t.Error("ParseWorkoutToolResult with garbage should not error, just return zero values")
	}
	if workoutID != 0 {
		t.Errorf("workoutID = %d, want 0 for invalid header", workoutID)
	}
}