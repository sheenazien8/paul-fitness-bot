package main

import "testing"

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
	if !contains(result, "Bench Press") {
		t.Error("FormatWorkoutForDisplay missing exercise name")
	}
	if !contains(result, "1") {
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}