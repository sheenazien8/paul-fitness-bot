package main

import "math"

func CalculateWorkoutScore(durationMinutes int, calories int, satisfaction int) float64 {
	durationScore := math.Min(float64(durationMinutes)/60.0, 1.0) * 100
	calorieScore := math.Min(float64(calories)/500.0, 1.0) * 100
	satisfactionScore := float64(satisfaction) / 10.0 * 100

	score := (durationScore * 0.3) + (calorieScore * 0.3) + (satisfactionScore * 0.4)
	score = math.Max(0, math.Min(100, score))

	return math.Round(score*10) / 10
}

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

func GetWeightChangeEmoji(change float64) string {
	switch {
	case change < -0.5:
		return "🎉"
	case change < 0:
		return "📉"
	case change == 0:
		return "➡️"
	case change > 0.5:
		return "📈"
	default:
		return "📊"
	}
}