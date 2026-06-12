package main

import "math"

// CalculateWorkoutScore computes a 0-100 score based on duration, calories, and satisfaction
// Formula: (duration_score * 0.3) + (calorie_score * 0.3) + (satisfaction_score * 0.4)
func CalculateWorkoutScore(durationMinutes int, calories int, satisfaction int) float64 {
	// Duration score: capped at 60 minutes for max score
	// 0-20min = low, 20-40min = moderate, 40-60min = high, 60+ = max
	durationScore := math.Min(float64(durationMinutes)/60.0, 1.0) * 100

	// Calorie score: capped at 500 calories for max score
	// Typical home workout: 200-400 cal
	calorieScore := math.Min(float64(calories)/500.0, 1.0) * 100

	// Satisfaction score: direct 1-10 mapping to 0-100
	satisfactionScore := float64(satisfaction) / 10.0 * 100

	// Weighted average
	score := (durationScore * 0.3) + (calorieScore * 0.3) + (satisfactionScore * 0.4)

	// Clamp to 0-100
	score = math.Max(0, math.Min(100, score))

	// Round to 1 decimal
	return math.Round(score*10) / 10
}

// GetScoreEmoji returns an emoji based on the workout score
func GetScoreEmoji(score float64) string {
	switch {
	case score >= 90:
		return "🔥"
	case score >= 75:
		return "💪"
	case score >= 60:
		return "👍"
	case score >= 40:
		return "🙂"
	default:
		return "😅"
	}
}

// GetScoreDescription returns a description based on the workout score
func GetScoreDescription(score float64) string {
	switch {
	case score >= 90:
		return "Luar biasa! Kamu memberikan yang terbaik!"
	case score >= 75:
		return "Hebat! Latihan yang solid!"
	case score >= 60:
		return "Bagus! Terus pertahankan!"
	case score >= 40:
		return "Lumayan, semangat lagi ya!"
	default:
		return "Yuk semangat! Setiap langih berarti!"
	}
}

// GetBMIStatus returns a status string for a given BMI
func GetBMIStatus(bmi float64) string {
	switch {
	case bmi < 18.5:
		return "Kurus"
	case bmi < 25:
		return "Normal"
	case bmi < 30:
		return "Gemuk"
	default:
		return "Obesitas"
	}
}

// GetBMIEmoji returns an emoji for a given BMI
func GetBMIEmoji(bmi float64) string {
	switch {
	case bmi < 18.5:
		return "📉"
	case bmi < 25:
		return "✅"
	case bmi < 30:
		return "⚠️"
	default:
		return "🔴"
	}
}

// GetWeightChangeEmoji returns an emoji based on weight change direction
// Positive change = gained weight (bad for fat loss goal)
// Negative change = lost weight (good for fat loss goal)
func GetWeightChangeEmoji(change float64) string {
	switch {
	case change < -0.5:
		return "🎉" // Lost significant weight
	case change < 0:
		return "📉" // Lost some weight
	case change == 0:
		return "➡️" // No change
	case change > 0.5:
		return "📈" // Gained significant weight
	default:
		return "📊" // Gained a little
	}
}