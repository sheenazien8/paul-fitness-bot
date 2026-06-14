package main

import (
	"fmt"
	"log/slog"
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
func (app *BotApp) SendTyping(chatID int64) {
	typing := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	app.Bot.Request(typing)
}

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

	app.ensureUser(msg.From)

	if msg.IsCommand() && msg.Command() == "start" {
		app.cmdStart(userID, chatID)
		return
	}

	if msg.Text != "" {
		app.SendTyping(chatID)
		response := ChatWithLLM(userID, msg.Text)

		workoutID := app.extractWorkoutIDFromResponse(userID, response)
		if workoutID > 0 {
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("✅ Selesai Latihan", fmt.Sprintf("workout_done:%d", workoutID)),
				),
			)
			app.SendMessageWithKeyboard(chatID, response, keyboard)
		} else {
			app.SendMessage(chatID, response)
		}
	}
}

// extractWorkoutIDFromResponse checks if there's a pending workout for today and returns its ID
func (app *BotApp) extractWorkoutIDFromResponse(userID int64, response string) int64 {
	now := time.Now()
	dayOfWeek := int(now.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7
	}

	workout, err := GetTodaysWorkout(userID, dayOfWeek)
	if err != nil || workout == nil {
		return 0
	}

	if strings.Contains(response, "Latihan Utama") || strings.Contains(response, "Latihan") || strings.Contains(response, "🏋️") {
		return workout.ID
	}
	return 0
}

// ensureUser creates user in DB if not exists
func (app *BotApp) ensureUser(tgUser *tgbotapi.User) {
	if _, err := GetUser(int64(tgUser.ID)); err == nil {
		return
	}

	user := &User{
		UserID:           int64(tgUser.ID),
		Username:         tgUser.UserName,
		FirstName:        tgUser.FirstName,
		Weight:           70,
		Height:           170,
		TargetWeight:     65,
		WorkoutDays:      "1,2,4,5",
		NotificationHour: 7,
	}

	if err := CreateUser(user); err != nil {
		slog.Error("create user failed", "user_id", tgUser.ID, "error", err)
	}
}

// cmdStart handles /start command
func (app *BotApp) cmdStart(userID, chatID int64) {
	app.SendTyping(chatID)
	response := ChatWithLLM(userID, "/start")
	app.SendMessage(chatID, response)
}

func (app *BotApp) handleCallback(callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	chatID := callback.Message.Chat.ID
	data := callback.Data

	slog.Info("callback received", "user_id", userID, "data", data)

	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	if _, err := app.Bot.Request(callbackConfig); err != nil {
		slog.Error("answer callback failed", "error", err)
	}

	parts := strings.SplitN(data, ":", 2)
	action := parts[0]
	param := ""
	if len(parts) > 1 {
		param = parts[1]
	}

	switch action {
	case "workout_done":
		app.handleWorkoutDoneCallback(userID, chatID, param)
	case "weight_confirm":
		app.handleWeightConfirmCallback(userID, chatID, param)
	default:
		app.SendTyping(chatID)
		response := ChatWithLLM(userID, data)
		app.SendMessage(chatID, response)
	}
}

// handleWorkoutDoneCallback handles "✅ Selesai Latihan" button
func (app *BotApp) handleWorkoutDoneCallback(userID, chatID int64, workoutIDStr string) {
	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		app.SendMessage(chatID, "❌ Workout tidak valid.")
		return
	}

	if err := UpdateSession(userID, "awaiting_workout_log", workoutID); err != nil {
		slog.Error("update session failed", "error", err)
	}

	app.SendMessage(chatID, "💪 Mantap! Berapa lama kamu latihan tadi? (dalam menit)\n\nContoh: <b>45</b>")
}

// handleWeightConfirmCallback handles weight confirmation button
func (app *BotApp) handleWeightConfirmCallback(userID, chatID int64, weightStr string) {
	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		app.SendMessage(chatID, "❌ Berat badan tidak valid.")
		return
	}

	user, _ := GetUser(userID)
	bmi := CalculateBMI(weight, user.Height)
	if err := SaveWeightLog(userID, weight, bmi); err != nil {
		app.SendMessage(chatID, "❌ Gagal menyimpan berat badan.")
		return
	}

	weightChange := weight - user.Weight
	changeEmoji := GetWeightChangeEmoji(weightChange)
	remaining := weight - user.TargetWeight

	bmiStatus := GetBMIStatus(bmi)
	bmiEmoji := GetBMIEmoji(bmi)

	text := fmt.Sprintf(
		"✅ <b>Berat badan dicatat!</b>\n\n"+
			"⚖️ %.1f kg\n"+
			"📏 BMI: %.1f %s %s\n"+
			"📊 Perubahan: %+.1f kg %s\n"+
			"🎯 Sisa ke target: %.1f kg\n\n"+
			"Terus konsisten ya! 💪",
		weight, bmi, bmiStatus, bmiEmoji, weightChange, changeEmoji, remaining,
	)

	app.SendMessage(chatID, text)
	_ = UpdateSession(userID, "idle", 0)
}

// HandleSessionInput processes multi-step session input for workout logging.
func (app *BotApp) HandleSessionInput(userID, chatID int64, text string) bool {
	session, err := GetSession(userID)
	if err != nil || session == nil {
		return false
	}

	switch session.State {
	case "awaiting_workout_log":
		app.handleWorkoutLogInput(userID, chatID, text, session)
		return true
	default:
		return false
	}
}

func (app *BotApp) handleWorkoutLogInput(userID, chatID int64, text string, session *UserSession) {
	switch session.State {
	case "awaiting_workout_log":
		duration, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || duration <= 0 || duration > 300 {
			app.SendMessage(chatID, "❌ Durasi tidak valid. Masukkan angka dalam menit.\nContoh: <b>45</b>")
			return
		}

		UpdateSession(userID, fmt.Sprintf("awaiting_calories:%d:%d", duration, session.WorkoutID), 0)
		app.SendMessage(chatID, fmt.Sprintf("✅ Durasi: %d menit\n\n🔥 Berapa kalori yang terbakar? (estimasi)\n\nContoh: <b>350</b>", duration))

	case "awaiting_calories":
		parts := strings.SplitN(session.State, ":", 3)
		if len(parts) < 3 {
			UpdateSession(userID, "idle", 0)
			response := ChatWithLLM(userID, text)
			app.SendMessage(chatID, response)
			return
		}

		duration, err := strconv.Atoi(parts[1])
		if err != nil {
			UpdateSession(userID, "idle", 0)
			response := ChatWithLLM(userID, text)
			app.SendMessage(chatID, response)
			return
		}

		calories, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || calories < 0 || calories > 5000 {
			app.SendMessage(chatID, "❌ Kalori tidak valid. Masukkan angka.\nContoh: <b>350</b>")
			return
		}

		workoutID := session.WorkoutID
		UpdateSession(userID, fmt.Sprintf("awaiting_satisfaction:%d:%d:%d", duration, calories, workoutID), 0)

		app.SendMessage(chatID, fmt.Sprintf("✅ Kalori: %d\n\n😊 Seberapa puas latihannya? (1-10)\n\n1 = Sangat tidak puas\n10 = Sangat puas", calories))

	case "awaiting_satisfaction":
		parts := strings.SplitN(session.State, ":", 4)
		if len(parts) < 4 {
			UpdateSession(userID, "idle", 0)
			response := ChatWithLLM(userID, text)
			app.SendMessage(chatID, response)
			return
		}

		duration, _ := strconv.Atoi(parts[1])
		calories, _ := strconv.Atoi(parts[2])
		workoutID, _ := strconv.ParseInt(parts[3], 10, 64)

		satisfaction, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || satisfaction < 1 || satisfaction > 10 {
			app.SendMessage(chatID, "❌ Rating tidak valid. Masukkan angka 1-10.\nContoh: <b>7</b>")
			return
		}

		score := CalculateWorkoutScore(duration, calories, satisfaction)

		log := &WorkoutLog{
			UserID:           userID,
			WorkoutID:        workoutID,
			DurationMinutes:  duration,
			Calories:         calories,
			Satisfaction:     satisfaction,
			Score:            score,
		}

		if err := SaveWorkoutLog(log); err != nil {
			slog.Error("save workout log failed", "error", err)
			app.SendMessage(chatID, "❌ Gagal menyimpan log workout.")
			UpdateSession(userID, "idle", 0)
			return
		}

		today := time.Now().Format("2006-01-02")
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
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

		UpdateSession(userID, "idle", 0)

		scoreEmoji := GetScoreEmoji(score)
		scoreDesc := GetScoreDescription(score)

		text := fmt.Sprintf(
			"🎉 <b>Workout selesai!</b>\n\n"+
				"⏱ Durasi: %d menit\n"+
				"🔥 Kalori: %d\n"+
				"😊 Puas: %d/10\n\n"+
				"📊 Skor: %.1f %s\n%s\n\n"+
				"💪 Streak: %d hari berturut-turut!\n\n"+
				"Istirahat yang cukup ya! 🔥",
			duration, calories, satisfaction, score, scoreEmoji, scoreDesc, newStreak,
		)

		app.SendMessage(chatID, text)
	}
}

// SendWorkoutNotification sends the daily workout to a user via LLM.
func (app *BotApp) SendWorkoutNotification(userID int64) {
	now := time.Now()
	dayOfWeek := int(now.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7
	}

	workoutType := GetWorkoutType(dayOfWeek)
	dayName := DayNames[dayOfWeek]

	var prompt string
	if workoutType == "" {
		prompt = "Hari ini bukan hari latihan. Kasih semangat aja buat istirahat dan recovery."
	} else {
		typeName := WorkoutTypeNames[workoutType]
		prompt = fmt.Sprintf("Sudah pagi! Hari ini %s, waktunya latihan %s. Generate workout untuk hari ini ya!", dayName, typeName)
	}

	response := ChatWithLLM(userID, prompt)

	workoutID := app.extractWorkoutIDFromResponse(userID, response)
	if workoutID > 0 {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Selesai Latihan", fmt.Sprintf("workout_done:%d", workoutID)),
			),
		)
		app.SendMessageWithKeyboard(userID, response, keyboard)
	} else {
		app.SendMessage(userID, response)
	}
}

// SendWeightReminder sends a weight reminder via LLM
func (app *BotApp) SendWeightReminder(userID int64) {
	prompt := "Waktunya timbang! Kirim berat badanmu hari ini ya. Ketik angkanya saja, contoh: 71.5"
	response := ChatWithLLM(userID, prompt)
	app.SendMessage(userID, response)
}

func parseWorkoutDays(days string) []int {
	var result []int
	for _, s := range strings.Split(days, ",") {
		s = strings.TrimSpace(s)
		if n, err := strconv.Atoi(s); err == nil {
			result = append(result, n)
		}
	}
	return result
}