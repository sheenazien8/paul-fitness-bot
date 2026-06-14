# 💪 Paul — Your Personal Fitness Assistant

AI-powered Telegram bot yang bantu kamu latihan di rumah. Chat langsung, workout auto-generate, progress otomatis terlacak.

## Fitur

- 🏋️ **AI Chat** — Ngobrol soal fitness, tanya apa aja tentang workout & nutrisi
- 🤖 **Auto Workout** — Latihan di-generate AI sesuai goal, level, dan alat kamu
- ⚖️ **Weight Tracking** — Catat berat badan, lihat progress ke target
- 📊 **Skor & Stats** — Skor 0-100 per latihan, streak, BMI, weekly report
- 😊 **Mood & Energy** — Log mood sebelum workout, affect intensity suggestion
- ⏰ **Smart Reminder** — Notifikasi latihan & timbangan otomatis
- 📋 **Weekly Report** — Summary mingguan otomatis (bisa dimatikan via settings)
- 🔄 **Exercise Tips** — Tanya form & tips buat exercise spesifik

## Cara Pakai

1. Chat [@paul_gym_instructor_bot](https://t.me/paul_gym_instructor_bot) di Telegram
2. `/start` — Paul bantu setup profil (goal, level, alat, jadwal)
3. Setelah onboarding, tinggal chat natural — "hari ini latihan apa?", "aku 71.5 kg", "lagi capek nih"
4. Klik **✅ Selesai Latihan** setelah workout, isi durasi & kalori

## Perintah

| Perintah | Deskripsi |
|----------|-----------|
| `/start` | Setup profil & onboarding |
| `/reset` | Reset chat history |

Selebihnya tinggal chat natural — Paul ngerti konteks dan data kamu.

## Skor Latihan

Formula: `(durasi × 0.3) + (kalori × 0.3) + (kepuasan × 0.4)`

Kepuasan punya bobot terbesar — karena konsisten & menikmati latihan lebih penting dari sekadar angka.

## Alat yang Didukung

- Dumbbell
- Resistance band
- Pull-up bar

## Teknologi

- **Go** — Bahasa utama
- **SQLite** (modernc.org/sqlite) — Database, pure Go
- **Ollama** — LLM engine (glm-5.1:cloud)
- **go-telegram-bot-api/v5** — Telegram Bot API
- **robfig/cron/v3** — Scheduling

---

Made by [Sheena](https://sheenazien.me) ⚡