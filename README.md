# 💪 Workout Bot

Telegram bot untuk latihan di rumah, dirancang untuk program diet/fat loss.

## Fitur

- 🏋️ **Jadwal Latihan Harian** — Push/Pull/Legs/Full Body split (Sen, Sel, Kam, Jum)
- ⚖️ **Tracking Berat Badan** — Catat berat setiap pagi, lihat progress ke target
- 📊 **Skor Latihan** — Skor 0-100 berdasarkan durasi, kalori, dan kepuasan
- 📈 **Statistik & Progress** — BMI, streak, weekly summary
- ⏰ **Reminder Otomatis** — Notifikasi pagi untuk latihan dan timbangan

## Persyaratan

- Go 1.23+
- SQLite (modernc.org/sqlite - pure Go, no CGO needed)

## Setup

### 1. Buat Bot Token

Chat dengan [@BotFather](https://t.me/BotFather) di Telegram:
1. `/newbot`
2. Pilih nama: `Workout Bot`
3. Pilih username: `your_username_bot`
4. Copy token yang diberikan

### 2. Konfigurasi

```bash
cp .env.example .env
# Edit .env, masukkan WORKOUT_BOT_TOKEN
```

### 3. Build & Jalankan

```bash
go build -o workout-bot .
./workout-bot
```

### 4. Install sebagai systemd Service

```bash
sudo cp workout-bot.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable workout-bot
sudo systemctl start workout-bot
```

## Perintah Bot

| Perintah | Deskripsi |
|----------|-----------|
| `/start` | Pesan selamat datang & profil |
| `/workout` | Lihat latihan hari ini |
| `/weight` | Catat berat badan |
| `/stats` | Statistik & progress |
| `/history` | Riwayat latihan terakhir |
| `/settings` | Ubah hari latihan & jam notifikasi |

## Cara Pakai

### Latihan
1. Bot mengirim menu latihan setiap pagi di hari latihan
2. Setelah selesai, klik **✅ Selesai Latihan**
3. Balas dengan format: `menit, kalori, tingkat puas (1-10)`
   - Contoh: `45, 320, 7`

### Berat Badan
1. Bot mengirim reminder pagi: "☀️ Pagi! Masukkan berat badanmu..."
2. Balas dengan angka saja: `71.5`
3. Bot menampilkan progress ke target 65kg

## Skor Latihan

Formula: `(durasi_score × 0.3) + (kalori_score × 0.3) + (kepuasan_score × 0.4)`

- **Durasi**: Maksimal di 60 menit (lebih lama tidak menambah skor)
- **Kalori**: Maksimal di 500 kalori
- **Kepuasan**: 1-10 langsung dikonversi ke 0-100

Bobot terbesar ada di kepuasan (40%) — karena konsisten dan menikmati latihan lebih penting dari sekadar angka.

## Alat Latihan

- Dumbbell
- Resistance band
- Pull-up bar

## Jadwal Split

| Hari | Tipe | Fokus |
|------|------|-------|
| Senin | Push | Dada, Bahu, Tricep |
| Selasa | Legs | Kaki, Glute |
| Rabu | Rest | — |
| Kamis | Pull | Punggung, Bicep |
| Jumat | Full Body | Seluruh tubuh |
| Sabtu-Minggu | Rest | — |

## Profil Default

- Berat: 72kg
- Tinggi: 167cm
- Target: 65kg
- BMI: ~25.8 (Gemuk)

## Teknologi

- **Go** — Bahasa utama
- **SQLite** (modernc.org/sqlite) — Database, pure Go tanpa CGO
- **go-telegram-bot-api/v5** — Telegram Bot API
- **robfig/cron/v3** — Penjadwalan
- **log/slog** — Structured logging