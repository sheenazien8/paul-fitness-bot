package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotApp holds the bot and its state
type BotApp struct {
	Bot *tgbotapi.BotAPI
}

// NewBotApp creates a new bot application
func NewBotApp(token string) (*BotApp, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	bot.Debug = false
	slog.Info("bot authorized", "username", bot.Self.UserName)

	return &BotApp{Bot: bot}, nil
}

// SendMessage sends a text message to a chat
func (app *BotApp) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	if _, err := app.Bot.Send(msg); err != nil {
		slog.Error("send message failed", "chat_id", chatID, "error", err)
	}
}

// SendMessageWithKeyboard sends a text message with inline keyboard
func (app *BotApp) SendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard
	if _, err := app.Bot.Send(msg); err != nil {
		slog.Error("send message with keyboard failed", "chat_id", chatID, "error", err)
	}
}

// HandleUpdate processes an incoming Telegram update
func (app *BotApp) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		if update.CallbackQuery != nil {
			app.handleCallback(update.CallbackQuery)
		}
		return
	}

	msg := update.Message
	userID := msg.From.ID
	chatID := msg.Chat.ID

	slog.Info("received message", "user_id", userID, "chat_id", chatID, "text", msg.Text)

	// Ensure user exists in DB
	app.ensureUser(msg.From)

	// Handle commands
	if msg.IsCommand() {
		app.handleCommand(userID, chatID, msg.Command(), msg.MessageID)
		return
	}

	// Handle text input based on session state
	app.handleTextInput(userID, chatID, msg.Text)
}

// ensureUser creates user in DB if not exists
func (app *BotApp) ensureUser(tgUser *tgbotapi.User) {
	user := &User{
		UserID:           int64(tgUser.ID),
		Username:         tgUser.UserName,
		FirstName:        tgUser.FirstName,
		Weight:           72,
		Height:           167,
		TargetWeight:     65,
		WorkoutDays:      "1,2,4,5",
		NotificationHour: 7,
	}

	if err := CreateUser(user); err != nil {
		slog.Error("create user failed", "user_id", tgUser.ID, "error", err)
	}
}

// handleCommand processes bot commands
func (app *BotApp) handleCommand(userID, chatID int64, command string, messageID int) {
	switch command {
	case "start":
		app.cmdStart(userID, chatID)
	case "workout":
		app.cmdWorkout(userID, chatID)
	case "weight":
		app.cmdWeight(userID, chatID)
	case "stats":
		app.cmdStats(userID, chatID)
	case "history":
		app.cmdHistory(userID, chatID)
	case "profile":
		app.cmdProfile(userID, chatID)
	case "settings":
		app.cmdSettings(userID, chatID)
	default:
		app.SendMessage(chatID, "Perintah tidak dikenali. Ketik /help untuk melihat daftar perintah.")
	}
}

// cmdStart handles /start command — triggers onboarding if not done
func (app *BotApp) cmdStart(userID, chatID int64) {
	prefs, err := GetUserPreferences(userID)
	if err != nil {
		slog.Error("get preferences failed", "user_id", userID, "error", err)
		// Fall through to normal welcome
	}

	if !prefs.OnboardingDone {
		// Start onboarding flow
		app.startOnboarding(userID, chatID)
		return
	}

	// Normal welcome for returning users
	user, err := GetUser(userID)
	if err != nil {
		slog.Error("get user failed", "user_id", userID, "error", err)
		return
	}

	bmi := CalculateBMI(user.Weight, user.Height)

	text := fmt.Sprintf(
		"💪 <b>Selamat Datang di Workout Bot!</b>\n\n"+
			"Halo %s! Aku akan membantumu mencapai target berat badanmu.\n\n"+
			"📊 <b>Profil Kamu:</b>\n"+
			"• Berat: %.1f kg\n"+
			"• Tinggi: %.0f cm\n"+
			"• BMI: %.1f (%s %s)\n"+
			"• Target: %.1f kg (sisa %.1f kg)\n\n"+
			"🏋️ <b>Jadwal Latihan:</b>\n"+
			"• Senin: Push (Dada, Bahu, Tricep)\n"+
			"• Selasa: Legs (Kaki, Glute)\n"+
			"• Kamis: Pull (Punggung, Bicep)\n"+
			"• Jumat: Full Body\n\n"+
			"📋 <b>Perintah yang tersedia:</b>\n"+
			"/workout — Lihat latihan hari ini\n"+
			"/weight — Catat berat badan\n"+
			"/stats — Lihat progress & statistik\n"+
			"/history — Riwayat latihan\n"+
			"/profile — Lihat/edit profil & preferensi\n"+
			"/settings — Ubah pengaturan\n\n"+
			"Semangat! 💪🔥",
		user.FirstName, user.Weight, user.Height, bmi, GetBMIStatus(bmi), GetBMIEmoji(bmi),
		user.TargetWeight, user.Weight-user.TargetWeight,
	)

	app.SendMessage(chatID, text)
}

// ============================================================
// ONBOARDING FLOW
// ============================================================

// startOnboarding begins the onboarding flow for a new user
func (app *BotApp) startOnboarding(userID, chatID int64) {
	if err := UpdateSession(userID, "onboarding_goal", 0); err != nil {
		slog.Error("update session for onboarding failed", "error", err)
	}

	text := "👋 <b>Selamat Datang di Workout Bot!</b>\n\n" +
		"Aku akan bantu kamu latihan di rumah! Tapi dulu, kita perlu setup profil kamu ya.\n\n" +
		"<b>Step 1/8:</b> 🎯 Apa tujuanmu?"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔥 Diet/Turun Berat", "onb:goal:diet"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💪 Bangun Otot", "onb:goal:muscle"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏃 Fitness Umum", "onb:goal:fitness"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 Maintenance", "onb:goal:maintenance"),
		),
	)

	app.SendMessageWithKeyboard(chatID, text, keyboard)
}

// onboardingGoal handles the goal selection
func (app *BotApp) onboardingGoal(userID, chatID int64, goal string) {
	if err := UpdateUserPreferenceField(userID, "goal", goal); err != nil {
		slog.Error("save goal preference failed", "error", err)
	}

	goalLabels := map[string]string{
		"diet":        "🔥 Diet/Turun Berat",
		"muscle":      "💪 Bangun Otot",
		"fitness":     "🏃 Fitness Umum",
		"maintenance": "🎯 Maintenance",
	}
	label := goalLabels[goal]

	if err := UpdateSession(userID, "onboarding_experience", 0); err != nil {
		slog.Error("update session failed", "error", err)
	}

	text := fmt.Sprintf("✅ Tujuan: %s\n\n<b>Step 2/8:</b> 🏋️ Pengalaman latihanmu?", label)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟢 Pemula", "onb:exp:beginner"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟡 Menengah", "onb:exp:intermediate"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔴 Lanjutan", "onb:exp:advanced"),
		),
	)

	app.SendMessageWithKeyboard(chatID, text, keyboard)
}

// onboardingExperience handles the experience level selection
func (app *BotApp) onboardingExperience(userID, chatID int64, level string) {
	if err := UpdateUserPreferenceField(userID, "experience_level", level); err != nil {
		slog.Error("save experience preference failed", "error", err)
	}

	levelLabels := map[string]string{
		"beginner":     "🟢 Pemula",
		"intermediate": "🟡 Menengah",
		"advanced":     "🔴 Lanjutan",
	}
	label := levelLabels[level]

	if err := UpdateSession(userID, "onboarding_equipment", 0); err != nil {
		slog.Error("update session failed", "error", err)
	}

	// Get current prefs to show toggle state
	prefs, _ := GetUserPreferences(userID)

	text := fmt.Sprintf("✅ Level: %s\n\n<b>Step 3/8:</b> 🔧 Alat yang tersedia?\n\nYang punya centang, yang ga punya ga usah.", label)

	dumbbellLabel := "❌ Dumbbell"
	if prefs.HasDumbbell {
		dumbbellLabel = "✅ Dumbbell"
	}
	bandLabel := "❌ Resistance Band"
	if prefs.HasResistanceBand {
		bandLabel = "✅ Resistance Band"
	}
	barLabel := "❌ Pull-Up Bar"
	if prefs.HasPullupBar {
		barLabel = "✅ Pull-Up Bar"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(dumbbellLabel, "onb:equip:dumbbell"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(bandLabel, "onb:equip:band"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(barLabel, "onb:equip:bar"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➡️ Lanjut", "onb:equip:done"),
		),
	)

	app.SendMessageWithKeyboard(chatID, text, keyboard)
}

// onboardingToggleEquipment toggles one equipment item during onboarding
func (app *BotApp) onboardingToggleEquipment(userID, chatID int64, item string) {
	prefs, _ := GetUserPreferences(userID)

	switch item {
	case "dumbbell":
		prefs.HasDumbbell = !prefs.HasDumbbell
		UpdateUserPreferenceField(userID, "has_dumbbell", boolToInt(prefs.HasDumbbell))
	case "band":
		prefs.HasResistanceBand = !prefs.HasResistanceBand
		UpdateUserPreferenceField(userID, "has_resistance_band", boolToInt(prefs.HasResistanceBand))
	case "bar":
		prefs.HasPullupBar = !prefs.HasPullupBar
		UpdateUserPreferenceField(userID, "has_pullup_bar", boolToInt(prefs.HasPullupBar))
	}

	// Refresh the equipment keyboard
	dumbbellLabel := "❌ Dumbbell"
	if prefs.HasDumbbell {
		dumbbellLabel = "✅ Dumbbell"
	}
	bandLabel := "❌ Resistance Band"
	if prefs.HasResistanceBand {
		bandLabel = "✅ Resistance Band"
	}
	barLabel := "❌ Pull-Up Bar"
	if prefs.HasPullupBar {
		barLabel = "✅ Pull-Up Bar"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(dumbbellLabel, "onb:equip:dumbbell"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(bandLabel, "onb:equip:band"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(barLabel, "onb:equip:bar"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➡️ Lanjut", "onb:equip:done"),
		),
	)

	text := "<b>Step 3/8:</b> 🔧 Alat yang tersedia?\n\nYang punya centang, yang ga punya ga usah."
	app.SendMessageWithKeyboard(chatID, text, keyboard)
}

// onboardingEquipmentDone finishes equipment selection, moves to weight input
func (app *BotApp) onboardingEquipmentDone(userID, chatID int64) {
	if err := UpdateSession(userID, "onboarding_weight", 0); err != nil {
		slog.Error("update session failed", "error", err)
	}

	app.SendMessage(chatID, "✅ Alat tersimpan!\n\n<b>Step 4/8:</b> ⚖️ Berat badanmu sekarang? (kg)\n\nKetik angkanya saja, contoh: <b>72</b>")
}

// onboardingWeight handles weight text input
func (app *BotApp) onboardingWeight(userID, chatID int64, text string) {
	weight, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil || weight < 20 || weight > 300 {
		app.SendMessage(chatID, "❌ Berat badan tidak valid. Masukkan angka dalam kg.\nContoh: <b>72</b>")
		return
	}

	user, _ := GetUser(userID)
	if err := UpdateUserProfile(userID, weight, user.Height, user.TargetWeight); err != nil {
		slog.Error("update weight in onboarding failed", "error", err)
	}

	if err := UpdateSession(userID, "onboarding_height", 0); err != nil {
		slog.Error("update session failed", "error", err)
	}

	app.SendMessage(chatID, fmt.Sprintf("✅ Berat: %.1f kg\n\n<b>Step 5/8:</b> 📏 Tinggi badanmu? (cm)\n\nKetik angkanya saja, contoh: <b>167</b>", weight))
}

// onboardingHeight handles height text input
func (app *BotApp) onboardingHeight(userID, chatID int64, text string) {
	height, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil || height < 100 || height > 250 {
		app.SendMessage(chatID, "❌ Tinggi badan tidak valid. Masukkan angka dalam cm.\nContoh: <b>167</b>")
		return
	}

	user, _ := GetUser(userID)
	if err := UpdateUserProfile(userID, user.Weight, height, user.TargetWeight); err != nil {
		slog.Error("update height in onboarding failed", "error", err)
	}

	if err := UpdateSession(userID, "onboarding_target", 0); err != nil {
		slog.Error("update session failed", "error", err)
	}

	app.SendMessage(chatID, fmt.Sprintf("✅ Tinggi: %.0f cm\n\n<b>Step 6/8:</b> 🎯 Target berat badanmu? (kg)\n\nKetik angkanya saja, contoh: <b>65</b>", height))
}

// onboardingTarget handles target weight text input
func (app *BotApp) onboardingTarget(userID, chatID int64, text string) {
	target, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil || target < 30 || target > 200 {
		app.SendMessage(chatID, "❌ Target berat tidak valid. Masukkan angka dalam kg.\nContoh: <b>65</b>")
		return
	}

	user, _ := GetUser(userID)
	if err := UpdateUserProfile(userID, user.Weight, user.Height, target); err != nil {
		slog.Error("update target in onboarding failed", "error", err)
	}

	if err := UpdateSession(userID, "onboarding_days", 0); err != nil {
		slog.Error("update session failed", "error", err)
	}

	text_response := fmt.Sprintf("✅ Target: %.1f kg\n\n<b>Step 7/8:</b> 📅 Hari apa aja mau latihan?", target)

	var rows [][]tgbotapi.InlineKeyboardButton
	allDays := []struct {
		num  int
		name string
	}{
		{1, "Senin"}, {2, "Selasa"}, {3, "Rabu"}, {4, "Kamis"},
		{5, "Jumat"}, {6, "Sabtu"}, {7, "Minggu"},
	}

	for _, d := range allDays {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(d.name, fmt.Sprintf("onb:days:%d", d.num)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ Lanjut", "onb:days:done"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	app.SendMessageWithKeyboard(chatID, text_response, keyboard)
}

// onboardingToggleDay toggles a workout day during onboarding
func (app *BotApp) onboardingToggleDay(userID, chatID int64, dayStr string) {
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return
	}

	user, _ := GetUser(userID)
	currentDays := parseWorkoutDays(user.WorkoutDays)

	// Toggle the day
	found := false
	var newDays []int
	for _, d := range currentDays {
		if d == day {
			found = true
			continue // Remove it
		}
		newDays = append(newDays, d)
	}
	if !found {
		newDays = append(newDays, day)
	}

	// Sort
	for i := 0; i < len(newDays); i++ {
		for j := i + 1; j < len(newDays); j++ {
			if newDays[i] > newDays[j] {
				newDays[i], newDays[j] = newDays[j], newDays[i]
			}
		}
	}

	dayStrs := make([]string, len(newDays))
	for i, d := range newDays {
		dayStrs[i] = strconv.Itoa(d)
	}
	daysString := strings.Join(dayStrs, ",")
	if err := UpdateUserSettings(userID, daysString, user.NotificationHour); err != nil {
		slog.Error("update workout days in onboarding failed", "error", err)
	}

	// Refresh the day keyboard with checkmarks
	var rows [][]tgbotapi.InlineKeyboardButton
	allDays := []struct {
		num  int
		name string
	}{
		{1, "Senin"}, {2, "Selasa"}, {3, "Rabu"}, {4, "Kamis"},
		{5, "Jumat"}, {6, "Sabtu"}, {7, "Minggu"},
	}

	for _, d := range allDays {
		label := d.name
		for _, cd := range newDays {
			if cd == d.num {
				label = "✅ " + d.name
				break
			}
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("onb:days:%d", d.num)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ Lanjut", "onb:days:done"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	app.SendMessageWithKeyboard(chatID, "<b>Step 7/8:</b> 📅 Hari apa aja mau latihan?", keyboard)
}

// onboardingDaysDone finishes day selection, moves to notification time
func (app *BotApp) onboardingDaysDone(userID, chatID int64) {
	if err := UpdateSession(userID, "onboarding_time", 0); err != nil {
		slog.Error("update session failed", "error", err)
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	times := []int{5, 6, 7, 8, 9, 10}

	row := []tgbotapi.InlineKeyboardButton{}
	for _, t := range times {
		label := fmt.Sprintf("%02d:00", t)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("onb:time:%d", t)))
		if len(row) == 3 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	app.SendMessageWithKeyboard(chatID, "✅ Hari latihan tersimpan!\n\n<b>Step 8/8:</b> ⏰ Jam berapa mau dikirim workout?", keyboard)
}

// onboardingTime handles notification time selection, completes onboarding
func (app *BotApp) onboardingTime(userID, chatID int64, hourStr string) {
	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		return
	}

	user, _ := GetUser(userID)
	if err := UpdateUserSettings(userID, user.WorkoutDays, hour); err != nil {
		slog.Error("update notification hour in onboarding failed", "error", err)
	}

	// Mark onboarding as done
	if err := CompleteOnboarding(userID); err != nil {
		slog.Error("complete onboarding failed", "error", err)
	}

	// Reset session
	if err := UpdateSession(userID, "idle", 0); err != nil {
		slog.Error("reset session failed", "error", err)
	}

	// Update scheduler
	_ = UpdateUserSchedule(userID)

	// Show profile summary
	app.showOnboardingSummary(userID, chatID, hour)
}

// showOnboardingSummary displays the completed profile
func (app *BotApp) showOnboardingSummary(userID, chatID int64, notifHour int) {
	user, _ := GetUser(userID)
	prefs, _ := GetUserPreferences(userID)
	bmi := CalculateBMI(user.Weight, user.Height)

	goalLabels := map[string]string{
		"diet":        "🔥 Diet/Turun Berat",
		"muscle":      "💪 Bangun Otot",
		"fitness":     "🏃 Fitness Umum",
		"maintenance": "🎯 Maintenance",
	}
	levelLabels := map[string]string{
		"beginner":     "🟢 Pemula",
		"intermediate": "🟡 Menengah",
		"advanced":     "🔴 Lanjutan",
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

	dayNums := parseWorkoutDays(user.WorkoutDays)
	dayLabels := make([]string, len(dayNums))
	for i, d := range dayNums {
		dayLabels[i] = DayNames[d]
	}

	text := fmt.Sprintf(
		"🎉 <b>Setup Selesai!</b>\n\n"+
			"👤 <b>Profil Kamu:</b>\n\n"+
			"🎯 Tujuan: %s\n"+
			"🏋️ Level: %s\n"+
			"🔧 Alat: %s\n\n"+
			"⚖️ Berat: %.1f kg\n"+
			"📏 Tinggi: %.0f cm\n"+
			"📏 BMI: %.1f (%s %s)\n"+
			"🎯 Target: %.1f kg\n\n"+
			"📅 Hari latihan: %s\n"+
			"⏰ Notifikasi: %02d:00 WIB\n\n"+
			"Ketik /workout untuk mulai latihan pertama! 💪🔥",
		goalLabels[prefs.Goal],
		levelLabels[prefs.ExperienceLevel],
		strings.Join(equipItems, ", "),
		user.Weight, user.Height, bmi, GetBMIStatus(bmi), GetBMIEmoji(bmi),
		user.TargetWeight,
		strings.Join(dayLabels, ", "),
		notifHour,
	)

	app.SendMessage(chatID, text)
}

// ============================================================
// WORKOUT COMMAND (LLM-enhanced)
// ============================================================

// cmdWorkout handles /workout command
func (app *BotApp) cmdWorkout(userID, chatID int64) {
	now := time.Now()
	dayOfWeek := int(now.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7 // Sunday = 7 in our system
	}

	// Check if today is a workout day
	workoutType := GetWorkoutType(dayOfWeek)
	if workoutType == "" {
		dayNames := []string{}
		for d, name := range DayNames {
			if _, ok := DayWorkoutMap[d]; ok {
				dayNames = append(dayNames, name)
			}
		}
		text := fmt.Sprintf(
			"😴 Hari ini bukan hari latihan.\n\n"+
				"Jadwal latihanmu:\n%s\n\n"+
				"Istirahat juga penting untuk recovery! 💪",
			strings.Join(dayNames, "\n"),
		)
		app.SendMessage(chatID, text)
		return
	}

	// Check if we already generated a workout today
	existingWorkout, _ := GetTodaysWorkout(userID, dayOfWeek)
	var workoutID int64
	var exercises []Exercise
	var warmUp []Exercise
	var coolDown []Exercise

	if existingWorkout != nil {
		// Reuse today's workout (caching)
		workoutID = existingWorkout.ID
		if err := json.Unmarshal([]byte(existingWorkout.Exercises), &exercises); err != nil {
			slog.Error("unmarshal exercises failed", "error", err)
			return
		}
		warmUp = GetWarmUp(workoutType)
		coolDown = GetCoolDown(workoutType)
	} else {
		// Try LLM generation first
		llmWorkout := app.tryLLMWorkout(userID, workoutType)
		if llmWorkout != nil {
			// Use LLM-generated workout (including warmup/cooldown from LLM)
			exercises = llmWorkout.Main
			warmUp = llmWorkout.Warmup
			coolDown = llmWorkout.Cooldown
			slog.Info("using LLM-generated workout", "user_id", userID, "type", workoutType,
				"warmup", len(warmUp), "main", len(exercises), "cooldown", len(coolDown))
		} else {
			// Fall back to pre-programmed pool
			exercises = GetWorkoutForDay(dayOfWeek)
			warmUp = GetWarmUp(workoutType)
			coolDown = GetCoolDown(workoutType)
			slog.Info("using pre-programmed workout (LLM fallback)", "user_id", userID, "type", workoutType)
		}

		if len(exercises) == 0 {
			app.SendMessage(chatID, "Maaf, tidak ada latihan yang tersedia untuk hari ini.")
			return
		}

		var err error
		workoutID, err = SaveWorkout(userID, dayOfWeek, workoutType, exercises)
		if err != nil {
			slog.Error("save workout failed", "error", err)
			return
		}
	}

	// Format workout message
	typeName := WorkoutTypeNames[workoutType]
	dayName := DayNames[dayOfWeek]
	dateStr := now.Format("2 January 2006")

	text := fmt.Sprintf("🏋️ <b>Menu Latihan Hari %s, %s:</b>\n<b>%s</b>\n\n", dayName, dateStr, typeName)

	// Warm-up section
	if len(warmUp) > 0 {
		text += "🔥 <b>Pemanasan (5-7 menit):</b>\n\n"
		for _, ex := range warmUp {
			text += fmt.Sprintf("• <b>%s</b> = %s = %s\n", ex.Name, ex.HowTo, ex.Reps)
		}
		text += "\n"
	}

	// Main workout
	text += "💪 <b>Latihan Utama:</b>\n\n"
	for i, ex := range exercises {
		text += fmt.Sprintf("• <b>%s</b> = %s = %s\n", ex.Name, ex.HowTo, ex.Reps)
		if i < len(exercises)-1 {
			text += "\n"
		}
	}

	// Cool-down section
	if len(coolDown) > 0 {
		text += fmt.Sprintf("\n❄️ <b>Cooling Down (5-10 menit):</b>\n\n")
		for _, ex := range coolDown {
			text += fmt.Sprintf("• <b>%s</b> = %s = %s\n", ex.Name, ex.HowTo, ex.Reps)
		}
	}

	text += fmt.Sprintf("\n💪 Semangat latihan! Setelah selesai, klik tombol di bawah.")

	// Add "Selesai Latihan" button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Selesai Latihan", fmt.Sprintf("workout_done:%d", workoutID)),
		),
	)

	app.SendMessageWithKeyboard(chatID, text, keyboard)
}

// tryLLMWorkout attempts to generate a workout via Ollama, returns nil on failure
func (app *BotApp) tryLLMWorkout(userID int64, dayType string) *LLMWorkoutResponse {
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

// ============================================================
// WEIGHT, STATS, HISTORY, PROFILE, SETTINGS
// ============================================================

// cmdWeight handles /weight command
func (app *BotApp) cmdWeight(userID, chatID int64) {
	// Set session state to await weight input
	if err := UpdateSession(userID, "awaiting_weight", 0); err != nil {
		slog.Error("update session failed", "error", err)
	}

	text := "⚖️ Masukkan berat badanmu hari ini (dalam kg):\n\nKetik angkanya saja, contoh: <b>71.5</b>"
	app.SendMessage(chatID, text)
}

// cmdStats handles /stats command
func (app *BotApp) cmdStats(userID, chatID int64) {
	stats, err := GetUserStats(userID)
	if err != nil {
		slog.Error("get stats failed", "error", err)
		app.SendMessage(chatID, "Maaf, gagal mengambil statistik.")
		return
	}

	bmiStatus := GetBMIStatus(stats.CurrentBMI)
	bmiEmoji := GetBMIEmoji(stats.CurrentBMI)

	weightChangeStr := "-"
	weightChangeEmoji := ""
	if stats.LastWeightChange != 0 {
		weightChangeStr = fmt.Sprintf("%+.1f kg", stats.LastWeightChange)
		weightChangeEmoji = GetWeightChangeEmoji(stats.LastWeightChange)
	}

	// Calculate progress percentage
	totalToLose := 72.0 - stats.TargetWeight // Starting was 72
	progressPct := 0.0
	if totalToLose > 0 {
		lost := totalToLose - stats.WeightRemaining
		progressPct = (lost / totalToLose) * 100
		if progressPct < 0 {
			progressPct = 0
		}
		if progressPct > 100 {
			progressPct = 100
		}
	}

	// Progress bar
	barLen := 15
	filled := int(progressPct / 100 * float64(barLen))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)

	text := fmt.Sprintf(
		"📊 <b>Statistik & Progress</b>\n\n"+
			"⚖️ <b>Berat Badan:</b>\n"+
			"• Sekarang: %.1f kg\n"+
			"• Target: %.1f kg\n"+
			"• Sisa: %.1f kg\n"+
			"• Perubahan terakhir: %s %s\n\n"+
			"📏 <b>BMI:</b> %.1f %s %s\n\n"+
			"📈 <b>Progress ke Target:</b>\n"+
			"  [%s] %.0f%%\n\n"+
			"🏋️ <b>Latihan Minggu Ini:</b>\n"+
			"• Sesi: %d\n"+
			"• Rata-rata Skor: %.1f\n"+
			"• Total Kalori: %d\n\n"+
			"🔥 <b>Streak:</b> %d hari berturut-turut",
		stats.CurrentWeight, stats.TargetWeight, stats.WeightRemaining,
		weightChangeStr, weightChangeEmoji,
		stats.CurrentBMI, bmiStatus, bmiEmoji,
		bar, progressPct,
		stats.WeeklySessions, stats.WeeklyAvgScore, stats.WeeklyCalories,
		stats.Streak,
	)

	app.SendMessage(chatID, text)
}

// cmdHistory handles /history command
func (app *BotApp) cmdHistory(userID, chatID int64) {
	logs, err := GetRecentWorkoutLogs(userID, 10)
	if err != nil {
		slog.Error("get history failed", "error", err)
		app.SendMessage(chatID, "Maaf, gagal mengambil riwayat.")
		return
	}

	if len(logs) == 0 {
		app.SendMessage(chatID, "📋 Belum ada riwayat latihan. Mulai latihan dengan /workout!")
		return
	}

	text := "📋 <b>Riwayat Latihan Terakhir:</b>\n\n"
	for i, l := range logs {
		emoji := GetScoreEmoji(l.Score)
		text += fmt.Sprintf(
			"%d. %s %s\n   ⏱ %d menit | 🔥 %d kal | 😊 %d/10\n   Skor: %.1f %s\n",
			i+1,
			l.LoggedAt.Format("02 Jan 15:04"),
			emoji,
			l.DurationMinutes,
			l.Calories,
			l.Satisfaction,
			l.Score,
			GetScoreDescription(l.Score),
		)
		if i < len(logs)-1 {
			text += "\n"
		}
	}

	app.SendMessage(chatID, text)
}

// cmdProfile handles /profile command — shows preferences + allows editing
func (app *BotApp) cmdProfile(userID, chatID int64) {
	user, err := GetUser(userID)
	if err != nil {
		slog.Error("get user failed", "error", err)
		return
	}

	prefs, _ := GetUserPreferences(userID)
	bmi := CalculateBMI(user.Weight, user.Height)

	goalLabels := map[string]string{
		"diet":        "🔥 Diet/Turun Berat",
		"muscle":      "💪 Bangun Otot",
		"fitness":     "🏃 Fitness Umum",
		"maintenance": "🎯 Maintenance",
	}
	levelLabels := map[string]string{
		"beginner":     "🟢 Pemula",
		"intermediate": "🟡 Menengah",
		"advanced":     "🔴 Lanjutan",
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

	goalLabel := goalLabels[prefs.Goal]
	if goalLabel == "" {
		goalLabel = goalLabels["diet"]
	}
	levelLabel := levelLabels[prefs.ExperienceLevel]
	if levelLabel == "" {
		levelLabel = levelLabels["beginner"]
	}

	text := fmt.Sprintf(
		"👤 <b>Profil Kamu:</b>\n\n"+
			"🎯 Tujuan: %s\n"+
			"🏋️ Level: %s\n"+
			"🔧 Alat: %s\n\n"+
			"⚖️ Berat badan: %.1f kg\n"+
			"📏 Tinggi badan: %.0f cm\n"+
			"🎯 Target berat: %.1f kg\n"+
			"📏 BMI: %.1f (%s %s)\n\n"+
			"Pilih data yang ingin diubah:",
		goalLabel, levelLabel, strings.Join(equipItems, ", "),
		user.Weight, user.Height, user.TargetWeight, bmi, GetBMIStatus(bmi), GetBMIEmoji(bmi),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("🎯 Tujuan: %s", goalLabel), "profile:goal"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("🏋️ Level: %s", levelLabel), "profile:level"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔧 Alat", "profile:equipment"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("⚖️ Berat: %.1f kg", user.Weight), "profile:weight"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("📏 Tinggi: %.0f cm", user.Height), "profile:height"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("🎯 Target: %.1f kg", user.TargetWeight), "profile:target"),
		),
	)

	app.SendMessageWithKeyboard(chatID, text, keyboard)
}

// cmdSettings handles /settings command
func (app *BotApp) cmdSettings(userID, chatID int64) {
	user, err := GetUser(userID)
	if err != nil {
		slog.Error("get user failed", "error", err)
		return
	}

	// Parse workout days for display
	dayNums := parseWorkoutDays(user.WorkoutDays)
	dayLabels := make([]string, len(dayNums))
	for i, d := range dayNums {
		dayLabels[i] = DayNames[d]
	}

	text := fmt.Sprintf(
		"⚙️ <b>Pengaturan:</b>\n\n"+
			"📅 Hari latihan: %s\n"+
			"⏰ Jam notifikasi: %d:00 WIB\n\n"+
			"Pilih opsi di bawah untuk mengubah:",
		strings.Join(dayLabels, ", "),
		user.NotificationHour,
	)

	// Inline keyboard for settings
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Ubah Hari Latihan", "settings:days"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏰ Ubah Jam Notifikasi", "settings:time"),
		),
	)

	app.SendMessageWithKeyboard(chatID, text, keyboard)
}

// ============================================================
// TEXT INPUT HANDLING
// ============================================================

// handleTextInput processes non-command text based on session state
func (app *BotApp) handleTextInput(userID, chatID int64, text string) {
	session, err := GetSession(userID)
	if err != nil {
		slog.Error("get session failed", "error", err)
		return
	}

	switch session.State {
	case "awaiting_workout_log":
		app.processWorkoutLog(userID, chatID, text, session.WorkoutID)
	case "awaiting_weight":
		app.processWeightInput(userID, chatID, text)
	case "awaiting_profile_weight", "awaiting_profile_height", "awaiting_profile_target":
		app.processProfileInput(userID, chatID, text, session.State)
	case "onboarding_weight":
		app.onboardingWeight(userID, chatID, text)
	case "onboarding_height":
		app.onboardingHeight(userID, chatID, text)
	case "onboarding_target":
		app.onboardingTarget(userID, chatID, text)
	default:
		// Try to auto-detect input format
		// Workout log format: "45, 320, 7"
		if isWorkoutLogFormat(text) {
			// Get latest workout for this user
			workout, err := GetLatestWorkout(userID)
			if err != nil {
				slog.Error("get latest workout failed", "error", err)
				return
			}
			app.processWorkoutLog(userID, chatID, text, workout.ID)
			return
		}

		// Weight format: just a number like "71.5"
		if isWeightFormat(text) {
			app.processWeightInput(userID, chatID, text)
			return
		}

		// Unknown input
		app.SendMessage(chatID, "🤔 Aku tidak mengerti. Ketik /workout untuk latihan, /weight untuk catat berat, atau /stats untuk statistik.")
	}
}

// processWorkoutLog parses and saves workout log input
func (app *BotApp) processWorkoutLog(userID, chatID int64, text string, workoutID int64) {
	// Parse format: "menit, kalori, tingkat puas"
	parts := strings.Split(text, ",")
	if len(parts) != 3 {
		app.SendMessage(chatID, "❌ Format salah. Balas dengan format: <b>menit, kalori, tingkat puas (1-10)</b>\nContoh: <code>45, 320, 7</code>")
		return
	}

	duration, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		app.SendMessage(chatID, "❌ Durasi tidak valid. Masukkan angka menit.\nContoh: <code>45, 320, 7</code>")
		return
	}

	calories, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		app.SendMessage(chatID, "❌ Kalori tidak valid. Masukkan angka kalori.\nContoh: <code>45, 320, 7</code>")
		return
	}

	satisfaction, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil || satisfaction < 1 || satisfaction > 10 {
		app.SendMessage(chatID, "❌ Tingkat kepuasan harus 1-10.\nContoh: <code>45, 320, 7</code>")
		return
	}

	// Calculate score
	score := CalculateWorkoutScore(duration, calories, satisfaction)

	// Save log
	log := &WorkoutLog{
		UserID:          userID,
		WorkoutID:       workoutID,
		DurationMinutes: duration,
		Calories:        calories,
		Satisfaction:     satisfaction,
		Score:           score,
	}
	if err := SaveWorkoutLog(log); err != nil {
		slog.Error("save workout log failed", "error", err)
		app.SendMessage(chatID, "❌ Gagal menyimpan log latihan.")
		return
	}

	// Update streak
	app.updateStreak(userID)

	// Reset session
	if err := UpdateSession(userID, "idle", 0); err != nil {
		slog.Error("reset session failed", "error", err)
	}

	emoji := GetScoreEmoji(score)
	desc := GetScoreDescription(score)

	text_response := fmt.Sprintf(
		"✅ <b>Latihan Tercatat!</b>\n\n"+
			"⏱ Durasi: %d menit\n"+
			"🔥 Kalori: %d\n"+
			"😊 Kepuasan: %d/10\n\n"+
			"🏆 Skor: <b>%.1f</b> %s\n"+
			"%s",
		duration, calories, satisfaction, score, emoji, desc,
	)

	app.SendMessage(chatID, text_response)
}

// processWeightInput parses and saves weight input
func (app *BotApp) processWeightInput(userID, chatID int64, text string) {
	weight, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil || weight < 20 || weight > 300 {
		app.SendMessage(chatID, "❌ Berat badan tidak valid. Masukkan angka dalam kg.\nContoh: <b>71.5</b>")
		return
	}

	// Get user for height (BMI calculation)
	user, err := GetUser(userID)
	if err != nil {
		slog.Error("get user failed", "error", err)
		return
	}

	bmi := CalculateBMI(weight, user.Height)
	bmiStatus := GetBMIStatus(bmi)
	bmiEmoji := GetBMIEmoji(bmi)

	// Save weight log
	if err := SaveWeightLog(userID, weight, bmi); err != nil {
		slog.Error("save weight log failed", "error", err)
		app.SendMessage(chatID, "❌ Gagal menyimpan berat badan.")
		return
	}

	// Check previous weight for change
	var changeStr string
	var changeEmoji string
	prevLog, err := GetPreviousWeightLog(userID)
	if err == nil {
		change := weight - prevLog.Weight
		changeStr = fmt.Sprintf("%+.1f kg", change)
		changeEmoji = GetWeightChangeEmoji(change)
	} else {
		changeStr = "(pertama kali)"
		changeEmoji = "🆕"
	}

	remaining := weight - user.TargetWeight
	progressPct := 0.0
	totalToLose := 72.0 - user.TargetWeight
	if totalToLose > 0 {
		lost := totalToLose - remaining
		progressPct = (lost / totalToLose) * 100
		if progressPct < 0 {
			progressPct = 0
		}
		if progressPct > 100 {
			progressPct = 100
		}
	}

	barLen := 15
	filled := int(progressPct / 100 * float64(barLen))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)

	text_response := fmt.Sprintf(
		"✅ <b>Berat Badan Tercatat!</b>\n\n"+
			"⚖️ Berat: %.1f kg\n"+
			"📏 BMI: %.1f %s %s\n"+
			"📊 Perubahan: %s %s\n"+
			"🎯 Sisa ke target: %.1f kg\n\n"+
			"📈 Progress:\n  [%s] %.0f%%",
		weight, bmi, bmiStatus, bmiEmoji,
		changeStr, changeEmoji,
		remaining,
		bar, progressPct,
	)

	// Reset session
	if err := UpdateSession(userID, "idle", 0); err != nil {
		slog.Error("reset session failed", "error", err)
	}

	app.SendMessage(chatID, text_response)
}

// processProfileInput handles text input for profile editing
func (app *BotApp) processProfileInput(userID, chatID int64, text string, state string) {
	value, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil || value <= 0 {
		app.SendMessage(chatID, "❌ Angka tidak valid. Masukkan angka yang benar.")
		return
	}

	user, err := GetUser(userID)
	if err != nil {
		slog.Error("get user for profile update failed", "error", err)
		return
	}

	switch state {
	case "awaiting_profile_weight":
		if value < 20 || value > 300 {
			app.SendMessage(chatID, "❌ Berat badan harus antara 20-300 kg.")
			return
		}
		if err := UpdateUserProfile(userID, value, user.Height, user.TargetWeight); err != nil {
			slog.Error("update profile weight failed", "error", err)
			return
		}
		bmi := CalculateBMI(value, user.Height)
		app.SendMessage(chatID, fmt.Sprintf("✅ Berat badan diubah ke <b>%.1f kg</b>\n📏 BMI: %.1f %s %s", value, bmi, GetBMIStatus(bmi), GetBMIEmoji(bmi)))

	case "awaiting_profile_height":
		if value < 100 || value > 250 {
			app.SendMessage(chatID, "❌ Tinggi badan harus antara 100-250 cm.")
			return
		}
		if err := UpdateUserProfile(userID, user.Weight, value, user.TargetWeight); err != nil {
			slog.Error("update profile height failed", "error", err)
			return
		}
		bmi := CalculateBMI(user.Weight, value)
		app.SendMessage(chatID, fmt.Sprintf("✅ Tinggi badan diubah ke <b>%.0f cm</b>\n📏 BMI: %.1f %s %s", value, bmi, GetBMIStatus(bmi), GetBMIEmoji(bmi)))

	case "awaiting_profile_target":
		if value < 30 || value > 200 {
			app.SendMessage(chatID, "❌ Target berat harus antara 30-200 kg.")
			return
		}
		if err := UpdateUserProfile(userID, user.Weight, user.Height, value); err != nil {
			slog.Error("update profile target failed", "error", err)
			return
		}
		remaining := user.Weight - value
		app.SendMessage(chatID, fmt.Sprintf("✅ Target berat diubah ke <b>%.1f kg</b>\n🎯 Sisa ke target: %.1f kg", value, remaining))
	}

	// Reset session
	if err := UpdateSession(userID, "idle", 0); err != nil {
		slog.Error("reset session failed", "error", err)
	}
}

// ============================================================
// CALLBACK HANDLING
// ============================================================

// handleCallback processes inline keyboard callbacks
func (app *BotApp) handleCallback(callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	chatID := callback.Message.Chat.ID
	data := callback.Data

	slog.Info("callback received", "user_id", userID, "data", data)

	// Answer callback
	callbackResponse := tgbotapi.NewCallback(callback.ID, "")
	if _, err := app.Bot.Request(callbackResponse); err != nil {
		slog.Error("callback response failed", "error", err)
	}

	parts := strings.SplitN(data, ":", 2)
	if len(parts) < 2 {
		return
	}

	action := parts[0]
	value := parts[1]

	switch action {
	case "workout_done":
		workoutID, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			slog.Error("parse workout ID failed", "error", err)
			return
		}
		// Set session to await workout log
		if err := UpdateSession(userID, "awaiting_workout_log", workoutID); err != nil {
			slog.Error("update session failed", "error", err)
		}
		app.SendMessage(chatID, "🎉 <b>Bagus! Latihan selesai!</b>\n\nBalas dengan format: <b>menit, kalori, tingkat puas (1-10)</b>\nContoh: <code>45, 320, 7</code>")

	case "onb":
		app.handleOnboardingCallback(userID, chatID, value)

	case "settings":
		switch value {
		case "days":
			app.showDaySettings(userID, chatID)
		case "time":
			app.showTimeSettings(chatID)
		}

	case "profile":
		app.handleProfileCallback(userID, chatID, value)

	case "set_day":
		app.setWorkoutDays(userID, chatID, value)

	case "set_time":
		hour, err := strconv.Atoi(value)
		if err != nil {
			return
		}
		if err := UpdateUserSettings(userID, "", hour); err != nil {
			slog.Error("update settings failed", "error", err)
		}
		// Need to get current workout days
		user, _ := GetUser(userID)
		if err := UpdateUserSettings(userID, user.WorkoutDays, hour); err != nil {
			slog.Error("update settings failed", "error", err)
		}
		app.SendMessage(chatID, fmt.Sprintf("✅ Jam notifikasi diubah ke <b>%d:00 WIB</b>", hour))
		_ = UpdateUserSchedule(userID)

	case "pref_goal":
		app.handlePrefGoalCallback(userID, chatID, value)

	case "pref_level":
		app.handlePrefLevelCallback(userID, chatID, value)

	case "pref_equip":
		app.handlePrefEquipCallback(userID, chatID, value)
	}
}

// handleOnboardingCallback processes onboarding inline keyboard callbacks
func (app *BotApp) handleOnboardingCallback(userID, chatID int64, value string) {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) < 2 {
		return
	}

	step := parts[0]
	val := parts[1]

	switch step {
	case "goal":
		app.onboardingGoal(userID, chatID, val)
	case "exp":
		app.onboardingExperience(userID, chatID, val)
	case "equip":
		if val == "done" {
			app.onboardingEquipmentDone(userID, chatID)
		} else {
			app.onboardingToggleEquipment(userID, chatID, val)
		}
	case "days":
		if val == "done" {
			app.onboardingDaysDone(userID, chatID)
		} else {
			app.onboardingToggleDay(userID, chatID, val)
		}
	case "time":
		app.onboardingTime(userID, chatID, val)
	}
}

// handleProfileCallback handles profile edit callbacks (enhanced with goal/level/equipment)
func (app *BotApp) handleProfileCallback(userID, chatID int64, field string) {
	switch field {
	case "goal":
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔥 Diet/Turun Berat", "pref_goal:diet"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💪 Bangun Otot", "pref_goal:muscle"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏃 Fitness Umum", "pref_goal:fitness"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🎯 Maintenance", "pref_goal:maintenance"),
			),
		)
		app.SendMessageWithKeyboard(chatID, "🎯 <b>Pilih tujuan baru:</b>", keyboard)

	case "level":
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🟢 Pemula", "pref_level:beginner"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🟡 Menengah", "pref_level:intermediate"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔴 Lanjutan", "pref_level:advanced"),
			),
		)
		app.SendMessageWithKeyboard(chatID, "🏋️ <b>Pilih level baru:</b>", keyboard)

	case "equipment":
		prefs, _ := GetUserPreferences(userID)
		dumbbellLabel := "❌ Dumbbell"
		if prefs.HasDumbbell {
			dumbbellLabel = "✅ Dumbbell"
		}
		bandLabel := "❌ Resistance Band"
		if prefs.HasResistanceBand {
			bandLabel = "✅ Resistance Band"
		}
		barLabel := "❌ Pull-Up Bar"
		if prefs.HasPullupBar {
			barLabel = "✅ Pull-Up Bar"
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(dumbbellLabel, "pref_equip:dumbbell"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(bandLabel, "pref_equip:band"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(barLabel, "pref_equip:bar"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Selesai", "pref_equip:done"),
			),
		)
		app.SendMessageWithKeyboard(chatID, "🔧 <b>Toggle alat yang tersedia:</b>", keyboard)

	case "weight":
		if err := UpdateSession(userID, "awaiting_profile_weight", 0); err != nil {
			slog.Error("update session for profile failed", "error", err)
		}
		app.SendMessage(chatID, "⚖️ Masukkan berat badanmu saat ini (kg):\nContoh: <b>72</b>")

	case "height":
		if err := UpdateSession(userID, "awaiting_profile_height", 0); err != nil {
			slog.Error("update session for profile failed", "error", err)
		}
		app.SendMessage(chatID, "📏 Masukkan tinggi badanmu (cm):\nContoh: <b>167</b>")

	case "target":
		if err := UpdateSession(userID, "awaiting_profile_target", 0); err != nil {
			slog.Error("update session for profile failed", "error", err)
		}
		app.SendMessage(chatID, "🎯 Masukkan target berat badanmu (kg):\nContoh: <b>65</b>")
	}
}

// handlePrefGoalCallback handles goal change from profile
func (app *BotApp) handlePrefGoalCallback(userID, chatID int64, goal string) {
	if err := UpdateUserPreferenceField(userID, "goal", goal); err != nil {
		slog.Error("update goal failed", "error", err)
	}
	goalLabels := map[string]string{
		"diet": "🔥 Diet/Turun Berat", "muscle": "💪 Bangun Otot",
		"fitness": "🏃 Fitness Umum", "maintenance": "🎯 Maintenance",
	}
	label := goalLabels[goal]
	app.SendMessage(chatID, fmt.Sprintf("✅ Tujuan diubah ke <b>%s</b>", label))
}

// handlePrefLevelCallback handles experience level change from profile
func (app *BotApp) handlePrefLevelCallback(userID, chatID int64, level string) {
	if err := UpdateUserPreferenceField(userID, "experience_level", level); err != nil {
		slog.Error("update level failed", "error", err)
	}
	levelLabels := map[string]string{
		"beginner": "🟢 Pemula", "intermediate": "🟡 Menengah", "advanced": "🔴 Lanjutan",
	}
	label := levelLabels[level]
	app.SendMessage(chatID, fmt.Sprintf("✅ Level diubah ke <b>%s</b>", label))
}

// handlePrefEquipCallback handles equipment toggle from profile
func (app *BotApp) handlePrefEquipCallback(userID, chatID int64, item string) {
	if item == "done" {
		app.SendMessage(chatID, "✅ Perubahan alat tersimpan!")
		return
	}

	prefs, _ := GetUserPreferences(userID)

	switch item {
	case "dumbbell":
		prefs.HasDumbbell = !prefs.HasDumbbell
		UpdateUserPreferenceField(userID, "has_dumbbell", boolToInt(prefs.HasDumbbell))
	case "band":
		prefs.HasResistanceBand = !prefs.HasResistanceBand
		UpdateUserPreferenceField(userID, "has_resistance_band", boolToInt(prefs.HasResistanceBand))
	case "bar":
		prefs.HasPullupBar = !prefs.HasPullupBar
		UpdateUserPreferenceField(userID, "has_pullup_bar", boolToInt(prefs.HasPullupBar))
	}

	// Refresh the equipment keyboard
	dumbbellLabel := "❌ Dumbbell"
	if prefs.HasDumbbell {
		dumbbellLabel = "✅ Dumbbell"
	}
	bandLabel := "❌ Resistance Band"
	if prefs.HasResistanceBand {
		bandLabel = "✅ Resistance Band"
	}
	barLabel := "❌ Pull-Up Bar"
	if prefs.HasPullupBar {
		barLabel = "✅ Pull-Up Bar"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(dumbbellLabel, "pref_equip:dumbbell"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(bandLabel, "pref_equip:band"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(barLabel, "pref_equip:bar"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Selesai", "pref_equip:done"),
		),
	)

	app.SendMessageWithKeyboard(chatID, "🔧 <b>Toggle alat yang tersedia:</b>", keyboard)
}

// ============================================================
// SETTINGS HELPERS
// ============================================================

// showDaySettings shows inline keyboard for selecting workout days
func (app *BotApp) showDaySettings(userID, chatID int64) {
	user, _ := GetUser(userID)
	currentDays := parseWorkoutDays(user.WorkoutDays)

	// Build day selection keyboard
	var rows [][]tgbotapi.InlineKeyboardButton

	allDays := []struct {
		num  int
		name string
	}{
		{1, "Senin"}, {2, "Selasa"}, {3, "Rabu"}, {4, "Kamis"},
		{5, "Jumat"}, {6, "Sabtu"}, {7, "Minggu"},
	}

	for _, d := range allDays {
		label := d.name
		// Mark current days
		for _, cd := range currentDays {
			if cd == d.num {
				label = "✅ " + d.name
				break
			}
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("set_day:%d", d.num)),
		))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	app.SendMessageWithKeyboard(chatID, "📅 <b>Pilih Hari Latihan:</b>\n\nKlik hari untuk menambah/menghapus:", keyboard)
}

// showTimeSettings shows inline keyboard for selecting notification time
func (app *BotApp) showTimeSettings(chatID int64) {
	var rows [][]tgbotapi.InlineKeyboardButton
	times := []int{5, 6, 7, 8, 9, 10}

	row := []tgbotapi.InlineKeyboardButton{}
	for _, t := range times {
		label := fmt.Sprintf("%d:00", t)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("set_time:%d", t)))
		if len(row) == 3 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	app.SendMessageWithKeyboard(chatID, "⏰ <b>Pilih Jam Notifikasi:</b>", keyboard)
}

// setWorkoutDays toggles a workout day
func (app *BotApp) setWorkoutDays(userID, chatID int64, dayStr string) {
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return
	}

	user, _ := GetUser(userID)
	currentDays := parseWorkoutDays(user.WorkoutDays)

	// Toggle the day
	found := false
	var newDays []int
	for _, d := range currentDays {
		if d == day {
			found = true
			continue // Remove it
		}
		newDays = append(newDays, d)
	}
	if !found {
		newDays = append(newDays, day)
	}

	// Sort
	for i := 0; i < len(newDays); i++ {
		for j := i + 1; j < len(newDays); j++ {
			if newDays[i] > newDays[j] {
				newDays[i], newDays[j] = newDays[j], newDays[i]
			}
		}
	}

	// Convert to string
	dayStrs := make([]string, len(newDays))
	for i, d := range newDays {
		dayStrs[i] = strconv.Itoa(d)
	}
	daysString := strings.Join(dayStrs, ",")

	if err := UpdateUserSettings(userID, daysString, user.NotificationHour); err != nil {
		slog.Error("update settings failed", "error", err)
	}

	// Show updated selection
	dayLabels := make([]string, len(newDays))
	for i, d := range newDays {
		dayLabels[i] = DayNames[d]
	}

	app.SendMessage(chatID, fmt.Sprintf("✅ Hari latihan diperbarui: <b>%s</b>", strings.Join(dayLabels, ", ")))
	_ = UpdateUserSchedule(userID)
}

// updateStreak updates the user's workout streak
func (app *BotApp) updateStreak(userID int64) {
	user, err := GetUser(userID)
	if err != nil {
		slog.Error("get user for streak failed", "error", err)
		return
	}

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	var newStreak int
	if user.LastWorkoutDate == yesterday {
		newStreak = user.Streak + 1
	} else if user.LastWorkoutDate == today {
		// Already worked out today, keep streak
		newStreak = user.Streak
	} else {
		// Streak broken, restart at 1
		newStreak = 1
	}

	if err := UpdateUserStreak(userID, newStreak, today); err != nil {
		slog.Error("update streak failed", "error", err)
	}
}

// SendWorkoutNotification sends the daily workout to a user
func (app *BotApp) SendWorkoutNotification(userID int64) {
	app.cmdWorkout(userID, userID)
}

// SendWeightReminder sends the morning weight reminder
func (app *BotApp) SendWeightReminder(userID int64) {
	// Set session state
	if err := UpdateSession(userID, "awaiting_weight", 0); err != nil {
		slog.Error("update session for weight failed", "error", err)
	}

	app.SendMessage(userID, "☀️ <b>Pagi!</b> Masukkan berat badanmu hari ini (kg):\n\nKetik angkanya saja, contoh: <b>71.5</b>")
}

// isWorkoutLogFormat checks if text matches "number, number, number" format
func isWorkoutLogFormat(text string) bool {
	matched, _ := regexp.MatchString(`^\d+\s*,\s*\d+\s*,\s*\d+$`, strings.TrimSpace(text))
	return matched
}

// isWeightFormat checks if text is just a number (potentially with decimal)
func isWeightFormat(text string) bool {
	matched, _ := regexp.MatchString(`^\d+\.?\d*$`, strings.TrimSpace(text))
	return matched
}