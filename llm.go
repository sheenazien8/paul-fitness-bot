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
	"time"
)

const (
	defaultOllamaURL   = "http://localhost:11434"
	defaultOllamaModel = "glm-5.1:cloud"
	llmTimeout         = 30 * time.Second
)

// ollamaRequest is the request body for Ollama API
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

// ollamaResponse is the response body from Ollama API
type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// getOllamaURL returns the Ollama API URL from env or default
func getOllamaURL() string {
	if url := os.Getenv("OLLAMA_URL"); url != "" {
		return strings.TrimRight(url, "/")
	}
	return defaultOllamaURL
}

// getOllamaModel returns the model name from env or default
func getOllamaModel() string {
	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		return model
	}
	return defaultOllamaModel
}

// GenerateWorkoutWithLLM calls Ollama to generate a workout based on user context
func GenerateWorkoutWithLLM(user *User, prefs *UserPreferences, dayType string, recentLogs []WorkoutLog) (*LLMWorkoutResponse, error) {
	prompt := buildWorkoutPrompt(user, prefs, dayType, recentLogs)

	result, err := callOllama(prompt)
	if err != nil {
		return nil, fmt.Errorf("ollama call failed: %w", err)
	}

	// Parse the JSON response
	var llmResp LLMWorkoutResponse
	if err := json.Unmarshal([]byte(result), &llmResp); err != nil {
		slog.Error("failed to parse LLM JSON response", "error", err, "raw", truncateString(result, 500))
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}

	// Validate structure
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

// buildWorkoutPrompt constructs the prompt for Ollama
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

	// Build equipment list
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

	// Build history context
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

		// Add progression hint
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

// callOllama sends a prompt to the Ollama API and returns the response text
func callOllama(prompt string) (string, error) {
	url := getOllamaURL() + "/api/generate"
	model := getOllamaModel()

	reqBody := ollamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	slog.Info("calling Ollama", "url", url, "model", model, "prompt_len", len(prompt))

	client := &http.Client{Timeout: llmTimeout}
	resp, err := client.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}

// validateLLMResponse checks that the LLM response has valid structure
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

	// Verify each exercise has required fields
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

// truncateString truncates a string for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}