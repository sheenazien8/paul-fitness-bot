package main

import (
	"time"
)

type Exercise struct {
	Name   string `json:"name"`
	HowTo  string `json:"how_to"`
	Reps   string `json:"reps"`
	Muscle string `json:"muscle"`
}

var WorkoutPool = map[string][][]Exercise{
	"push": {
		{
			{Name: "Dumbbell Bench Press", HowTo: "Berbaring di bangku, pegang dumbbell di samping dada, dorong ke atas hingga lengan lurus", Reps: "3x12", Muscle: "Chest"},
			{Name: "Dumbbell Incline Press", HowTo: "Bangku miring 30-45 derajat, dorong dumbbell dari samping dada ke atas", Reps: "3x10", Muscle: "Upper Chest"},
			{Name: "Dumbbell Fly", HowTo: "Berbaring di bangku, lengan sedikit tekuk, buka lebar lalu rapatkan di atas dada", Reps: "3x12", Muscle: "Chest"},
			{Name: "Dumbbell Shoulder Press", HowTo: "Duduk, dumbbell di samping bahu, dorong vertikal ke atas", Reps: "3x10", Muscle: "Shoulders"},
			{Name: "Dumbbell Lateral Raise", HowTo: "Berdiri, angkat dumbbell ke samping hingga sejajar bahu", Reps: "3x15", Muscle: "Side Delts"},
			{Name: "Dumbbell Tricep Extension", HowTo: "Duduk, pegang dumbbell di belakang kepala, luruskan ke atas", Reps: "3x12", Muscle: "Triceps"},
			{Name: "Diamond Push-Up", HowTo: "Push-up dengan tangan rapat membentuk diamond, turun dan dorong", Reps: "3x10", Muscle: "Triceps"},
			{Name: "Dumbbell Front Raise", HowTo: "Berdiri, angkat dumbbell ke depan hingga sejajar bahu secara bergantian", Reps: "3x12", Muscle: "Front Delts"},
		},
		{
			{Name: "Dumbbell Floor Press", HowTo: "Berbaring di lantai, pegang dumbbell di samping dada, dorong ke atas", Reps: "4x10", Muscle: "Chest"},
			{Name: "Push-Up", HowTo: "Tangan selebar bahu, turunkan dada ke lantai lalu dorong kembali", Reps: "3x15", Muscle: "Chest"},
			{Name: "Dumbbell Arnold Press", HowTo: "Duduk, mulai posisi curl, rotasi dan dorong ke atas seperti shoulder press", Reps: "3x10", Muscle: "Shoulders"},
			{Name: "Resistance Band Chest Press", HowTo: "Lingkarkan band di punggung, dorong tangan ke depan hingga lengan lurus", Reps: "3x12", Muscle: "Chest"},
			{Name: "Dumbbell Upright Row", HowTo: "Berdiri, tarik dumbbell naik sepanjang tubuh hingga sejajar dada", Reps: "3x12", Muscle: "Shoulders"},
			{Name: "Dumbbell Kickback", HowTo: "Miring tubuh, lengan tekuk 90°, luruskan ke belakang hingga penuh", Reps: "3x12", Muscle: "Triceps"},
			{Name: "Resistance Band Lateral Raise", HowTo: "Injak band, tarik pegangan ke samping hingga sejajar bahu", Reps: "3x15", Muscle: "Side Delts"},
			{Name: "Pike Push-Up", HowTo: "Push-up dengan pinggul tinggi, turunkan kepala ke lantai lalu dorong", Reps: "3x8", Muscle: "Shoulders"},
		},
		{
			{Name: "Dumbbell Decline Press", HowTo: "Berbaring di bangku miring ke bawah, dorong dumbbell dari samping dada ke atas", Reps: "3x10", Muscle: "Lower Chest"},
			{Name: "Dumbbell Pullover", HowTo: "Berbaring, pegang dumbbell di atas dada, turunkan ke belakang kepala lalu tarik kembali", Reps: "3x12", Muscle: "Chest"},
			{Name: "Dumbbell Shoulder Press", HowTo: "Berdiri, dumbbell di samping bahu, dorong vertikal ke atas", Reps: "4x8", Muscle: "Shoulders"},
			{Name: "Resistance Band Front Raise", HowTo: "Injak band, tarik pegangan ke depan hingga sejajar bahu", Reps: "3x12", Muscle: "Front Delts"},
			{Name: "Dumbbell Reverse Fly", HowTo: "Miring tubuh 45°, buka lengan ke samping lalu rapatkan", Reps: "3x12", Muscle: "Rear Delts"},
			{Name: "Close-Grip Push-Up", HowTo: "Push-up dengan tangan selebar bahu, siku rapat ke tubuh", Reps: "3x12", Muscle: "Triceps"},
			{Name: "Dumbbell Overhead Tricep Extension", HowTo: "Duduk, pegang satu dumbbell dengan dua tangan di belakang kepala, dorong ke atas", Reps: "3x10", Muscle: "Triceps"},
			{Name: "Resistance Band Chest Fly", HowTo: "Lingkarkan band di punggung, buka lengan lebar lalu rapatkan di depan", Reps: "3x12", Muscle: "Chest"},
		},
		{
			{Name: "Dumbbell Bench Press", HowTo: "Berbaring di bangku, pegang dumbbell di samping dada, dorong ke atas", Reps: "4x10", Muscle: "Chest"},
			{Name: "Resistance Band Shoulder Press", HowTo: "Injak band, tarik pegangan dari bahu ke atas hingga lengan lurus", Reps: "3x12", Muscle: "Shoulders"},
			{Name: "Dumbbell Incline Fly", HowTo: "Bangku miring 30°, buka lengan lebar lalu rapatkan di atas dada", Reps: "3x10", Muscle: "Upper Chest"},
			{Name: "Dumbbell Lateral Raise", HowTo: "Berdiri, angkat dumbbell ke samping hingga sejajar bahu", Reps: "4x12", Muscle: "Side Delts"},
			{Name: "Bench Dip", HowTo: "Tangan di bangku di belakang, turunkan tubuh lalu dorong kembali", Reps: "3x12", Muscle: "Triceps"},
			{Name: "Dumbbell Tricep Extension", HowTo: "Berdiri, pegang dumbbell di belakang kepala, luruskan ke atas", Reps: "3x10", Muscle: "Triceps"},
			{Name: "Push-Up to Side Plank", HowTo: "Lakukan push-up, lalu rotasi ke side plank bergantian kiri dan kanan", Reps: "3x8", Muscle: "Chest/Core"},
		},
		{
			{Name: "Resistance Band Chest Press", HowTo: "Lingkarkan band di punggung, dorong tangan ke depan hingga lengan lurus", Reps: "4x12", Muscle: "Chest"},
			{Name: "Dumbbell Shoulder Press", HowTo: "Duduk, dumbbell di samping bahu, dorong vertikal ke atas", Reps: "3x10", Muscle: "Shoulders"},
			{Name: "Dumbbell Squeeze Press", HowTo: "Berbaring, dua dumbbell dirapatkan di dada, dorong ke atas", Reps: "3x12", Muscle: "Inner Chest"},
			{Name: "Dumbbell Front Raise", HowTo: "Berdiri, angkat dumbbell ke depan hingga sejajar bahu", Reps: "3x10", Muscle: "Front Delts"},
			{Name: "Dumbbell Overhead Extension", HowTo: "Dua tangan pegang satu dumbbell di belakang kepala, dorong ke atas", Reps: "3x12", Muscle: "Triceps"},
			{Name: "Resistance Band Lateral Raise", HowTo: "Injak band, tarik pegangan ke samping hingga sejajar bahu", Reps: "3x15", Muscle: "Side Delts"},
			{Name: "Tricep Dip on Floor", HowTo: "Duduk di lantai, tangan di belakang, angkat pinggul naik turun", Reps: "3x10", Muscle: "Triceps"},
		},
	},
	"pull": {
		{
			{Name: "Pull-Up", HowTo: "Pegang pull-up bar selebar bahu, tarik tubuh naik hingga dagu melewati bar", Reps: "3x8", Muscle: "Back"},
			{Name: "Dumbbell Row", HowTo: "Miring tubuh, tarik dumbbell dari bawah ke samping perut", Reps: "3x12", Muscle: "Back"},
			{Name: "Dumbbell Bicep Curl", HowTo: "Berdiri, curl dumbbell dari paha ke bahu, tangan tetap di samping", Reps: "3x12", Muscle: "Biceps"},
			{Name: "Resistance Band Row", HowTo: "Kaitkan band, tarik pegangan ke arah perut, siku tarik ke belakang", Reps: "3x12", Muscle: "Back"},
			{Name: "Dumbbell Reverse Fly", HowTo: "Miring tubuh, buka lengan ke samping hingga sejajar bahu", Reps: "3x12", Muscle: "Rear Delts"},
			{Name: "Dumbbell Hammer Curl", HowTo: "Berdiri, curl dumbbell dengan posisi tangan netral (telapak menghadap satu sama lain)", Reps: "3x10", Muscle: "Biceps"},
			{Name: "Resistance Band Pull-Apart", HowTo: "Pegang band di depan dada, tarik ke samping hingga band menyentuh dada", Reps: "3x15", Muscle: "Rear Delts"},
			{Name: "Chin-Up", HowTo: "Pull-up dengan telapak menghadap ke diri, tarik tubuh naik", Reps: "3x6", Muscle: "Biceps/Back"},
		},
		{
			{Name: "Dumbbell Renegade Row", HowTo: "Posisi plank, tarik dumbbell ke samping perut bergantian", Reps: "3x10", Muscle: "Back/Core"},
			{Name: "Resistance Band Lat Pulldown", HowTo: "Kaitkan band di atas, tarik pegangan ke bawah ke arah dada", Reps: "3x12", Muscle: "Lats"},
			{Name: "Dumbbell Concentration Curl", HowTo: "Duduk, siku di paha, curl dumbbell dengan satu tangan", Reps: "3x10", Muscle: "Biceps"},
			{Name: "Pull-Up (Wide Grip)", HowTo: "Pegang bar lebih lebar dari bahu, tarik tubuh naik", Reps: "3x6", Muscle: "Lats"},
			{Name: "Dumbbell Shrug", HowTo: "Berdiri, angkat bahu ke atas (shrug) tahan 2 detik, turunkan", Reps: "3x15", Muscle: "Traps"},
			{Name: "Resistance Band Bicep Curl", HowTo: "Injak band, curl pegangan ke arah bahu", Reps: "3x12", Muscle: "Biceps"},
			{Name: "Dumbbell Row (Two Arm)", HowTo: "Miring tubuh 45°, tarik kedua dumbbell ke samping perut", Reps: "3x10", Muscle: "Back"},
			{Name: "Resistance Band Face Pull", HowTo: "Kaitkan band setinggi wajah, tarik ke arah wajah siku ke belakang", Reps: "3x15", Muscle: "Rear Delts"},
		},
		{
			{Name: "Negative Pull-Up", HowTo: "Mulai dari posisi dagu di atas bar, turunkan perlahan selama 5 detik", Reps: "3x5", Muscle: "Back"},
			{Name: "Dumbbell Single Arm Row", HowTo: "Satu lutut di bangku, tarik dumbbell ke samping perut", Reps: "3x10", Muscle: "Back"},
			{Name: "Dumbbell Preacher Curl", HowTo: "Siku di permukaan miring, curl dumbbell ke atas", Reps: "3x10", Muscle: "Biceps"},
			{Name: "Resistance Band Seated Row", HowTo: "Duduk, kaki lurus, tarik band ke arah perut", Reps: "3x12", Muscle: "Back"},
			{Name: "Dumbbell Reverse Curl", HowTo: "Berdiri, pegang dumbbell telapak ke bawah, curl ke atas", Reps: "3x10", Muscle: "Forearms"},
			{Name: "Pull-Up (Close Grip)", HowTo: "Tangan rapat di pull-up bar, tarik tubuh naik", Reps: "3x6", Muscle: "Back/Biceps"},
			{Name: "Resistance Band Shrug", HowTo: "Injak band, pegang di samping, angkat bahu ke atas", Reps: "3x15", Muscle: "Traps"},
			{Name: "Dumbbell Zottman Curl", HowTo: "Curl telapak ke atas, di atas rotasi telapak ke bawah, turunkan perlahan", Reps: "3x10", Muscle: "Biceps/Forearms"},
		},
		{
			{Name: "Pull-Up", HowTo: "Pegang pull-up bar selebar bahu, tarik tubuh naik hingga dagu melewati bar", Reps: "4x6", Muscle: "Back"},
			{Name: "Dumbbell Bent-Over Row", HowTo: "Miring tubuh, tarik kedua dumbbell ke samping dada", Reps: "3x10", Muscle: "Upper Back"},
			{Name: "Resistance Band Curl", HowTo: "Injak band, curl pegangan ke bahu, tahan 1 detik di atas", Reps: "3x12", Muscle: "Biceps"},
			{Name: "Dumbbell Reverse Fly", HowTo: "Miring 45°, buka lengan lebar ke samping", Reps: "3x12", Muscle: "Rear Delts"},
			{Name: "Chin-Up", HowTo: "Pull-up telapak menghadap ke diri, fokus pada bicep", Reps: "3x8", Muscle: "Biceps"},
			{Name: "Dumbbell Hammer Curl", HowTo: "Posisi tangan netral, curl ke arah bahu", Reps: "3x10", Muscle: "Brachialis"},
			{Name: "Resistance Band Row (High)", HowTo: "Kaitkan band setinggi dada, tarik ke arah tubuh", Reps: "3x12", Muscle: "Upper Back"},
		},
		{
			{Name: "Resistance Band Lat Pulldown", HowTo: "Kaitkan band di atas kepala, tarik ke bawah ke dada, siku ke belakang", Reps: "4x12", Muscle: "Lats"},
			{Name: "Dumbbell Single Arm Row", HowTo: "Satu tangan di bangku, tarik dumbbell ke samping perut", Reps: "3x10", Muscle: "Back"},
			{Name: "Pull-Up", HowTo: "Pegang bar, tarik tubuh naik hingga dagu melewati bar", Reps: "3x8", Muscle: "Back"},
			{Name: "Dumbbell Bicep Curl 21s", HowTo: "7 repetisi setengah bawah, 7 setengah atas, 7 penuh", Reps: "3x21", Muscle: "Biceps"},
			{Name: "Dumbbell Shrug", HowTo: "Berdiri, angkat bahu ke atas tahan 2 detik, turunkan", Reps: "3x12", Muscle: "Traps"},
			{Name: "Resistance Band Face Pull", HowTo: "Kaitkan band setinggi wajah, tarik ke arah wajah", Reps: "3x15", Muscle: "Rear Delts"},
			{Name: "Negative Chin-Up", HowTo: "Mulai dagu di atas bar, turunkan perlahan 5 detik", Reps: "3x5", Muscle: "Back/Biceps"},
		},
	},
	"legs": {
		{
			{Name: "Dumbbell Goblet Squat", HowTo: "Pegang dumbbell di depan dada, squat hingga paha sejajar lantai", Reps: "3x12", Muscle: "Quads"},
			{Name: "Dumbbell Romanian Deadlift", HowTo: "Berdiri, dumbbell di depan paha, dorong pinggul ke belakang turunkan dumbbell", Reps: "3x10", Muscle: "Hamstrings"},
			{Name: "Dumbbell Lunge", HowTo: "Langkah maju, turunkan lutut belakang hingga hampir lantai, dorong kembali", Reps: "3x10/leg", Muscle: "Quads"},
			{Name: "Dumbbell Calf Raise", HowTo: "Berdiri di tepi bangku, angkat tumit naik turunkan perlahan", Reps: "3x15", Muscle: "Calves"},
			{Name: "Resistance Band Glute Bridge", HowTo: "Band di atas pinggul, berbaring, dorong pinggul naik", Reps: "3x12", Muscle: "Glutes"},
			{Name: "Dumbbell Step-Up", HowTo: "Naik ke bangku dengan satu kaki, dorong ke atas, turunkan kembali", Reps: "3x10/leg", Muscle: "Quads"},
			{Name: "Resistance Band Lateral Walk", HowTo: "Band di pergelangan kaki, jalan menyamping dalam posisi squat", Reps: "3x15/side", Muscle: "Glutes"},
			{Name: "Bodyweight Squat Jump", HowTo: "Squat lalu lompat ke atas, mendarat lembut kembali ke squat", Reps: "3x10", Muscle: "Quads"},
		},
		{
			{Name: "Dumbbell Front Squat", HowTo: "Dumbbell di bahu depan, squat hingga paha sejajar lantai", Reps: "3x10", Muscle: "Quads"},
			{Name: "Dumbbell Single Leg RDL", HowTo: "Berdiri satu kaki, miring tubuh ke depan, dumbbell turun sejajar kaki", Reps: "3x8/leg", Muscle: "Hamstrings"},
			{Name: "Dumbbell Bulgarian Split Squat", HowTo: "Kaki belakang di bangku, squat dengan kaki depan", Reps: "3x8/leg", Muscle: "Quads"},
			{Name: "Resistance Band Leg Curl", HowTo: "Berbaring telentang, band di pergelangan kaki, tekuk lutut ke arah pinggul", Reps: "3x12", Muscle: "Hamstrings"},
			{Name: "Dumbbell Calf Raise (Seated)", HowTo: "Duduk, dumbbell di lutut, angkat tumit naik turunkan", Reps: "3x15", Muscle: "Calves"},
			{Name: "Resistance Band Hip Thrust", HowTo: "Band di atas pinggul, punggung di bangku, dorong pinggul naik", Reps: "3x12", Muscle: "Glutes"},
			{Name: "Dumbbell Reverse Lunge", HowTo: "Langkah ke belakang, turunkan lutut, dorong kembali", Reps: "3x10/leg", Muscle: "Quads"},
		},
		{
			{Name: "Dumbbell Sumo Squat", HowTo: "Kaki lebar, ujung kaki keluar, squat sambil pegang dumbbell di bawah", Reps: "3x12", Muscle: "Inner Thighs"},
			{Name: "Dumbbell Deadlift", HowTo: "Berdiri, dumbbell di depan paha, pinggul dorong belakang lalu kembali", Reps: "3x10", Muscle: "Hamstrings"},
			{Name: "Dumbbell Walking Lunge", HowTo: "Langkah maju bergantian sambil pegang dumbbell di samping", Reps: "3x12/leg", Muscle: "Quads"},
			{Name: "Resistance Band Glute Kickback", HowTo: "Band di pergelangan kaki, tendang kaki ke belakang hingga lurus", Reps: "3x12/leg", Muscle: "Glutes"},
			{Name: "Dumbbell Calf Raise", HowTo: "Berdiri di tepi, angkat tumit naik tahan 2 detik turunkan", Reps: "4x15", Muscle: "Calves"},
			{Name: "Resistance Band Lateral Walk", HowTo: "Band di pergelangan, posisi squat, jalan menyamping", Reps: "3x20/side", Muscle: "Glutes"},
			{Name: "Wall Sit", HowTo: "Punggung di dinding, posisi squat 90°, tahan 30-60 detik", Reps: "3x45d", Muscle: "Quads"},
			{Name: "Dumbbell Step-Up", HowTo: "Naik bangku dengan dumbbell, kaki depan dorong, turun kembali", Reps: "3x10/leg", Muscle: "Glutes"},
		},
		{
			{Name: "Dumbbell Goblet Squat", HowTo: "Pegang dumbbell di dada, squat dalam hingga paha sejajar", Reps: "4x10", Muscle: "Quads"},
			{Name: "Dumbbell Romanian Deadlift", HowTo: "Kaki sedikit tekuk, dorong pinggul ke belakang, dumbbell turun sejajar kaki", Reps: "3x10", Muscle: "Hamstrings"},
			{Name: "Dumbbell Curtsy Lunge", HowTo: "Langkah satu kaki silang ke belakang, squat, kembali", Reps: "3x10/leg", Muscle: "Glutes"},
			{Name: "Resistance Band Leg Extension", HowTo: "Duduk, band di pergelangan kaki, luruskan kaki ke depan", Reps: "3x12", Muscle: "Quads"},
			{Name: "Dumbbell Single Leg Calf Raise", HowTo: "Satu kaki di tepi, angkat tumit naik turunkan perlahan", Reps: "3x12/leg", Muscle: "Calves"},
			{Name: "Resistance Band Hip Abduction", HowTo: "Berbaring miring, band di pergelangan, buka kaki ke atas", Reps: "3x12/side", Muscle: "Outer Thighs"},
			{Name: "Dumbbell Glute Bridge", HowTo: "Berbaring, dumbbell di pinggul, dorong pinggul naik", Reps: "3x12", Muscle: "Glutes"},
		},
		{
			{Name: "Dumbbell Split Squat", HowTo: "Kaki depan di depan, lutut belakang hampir lantai, dorong kembali", Reps: "3x10/leg", Muscle: "Quads"},
			{Name: "Dumbbell RDL", HowTo: "Pinggul dorong ke belakang, dumbbell turun sejajar kaki, berdiri kembali", Reps: "4x10", Muscle: "Hamstrings"},
			{Name: "Resistance Band Squat", HowTo: "Band di bawah kaki, pegang di bahu, squat dalam", Reps: "3x15", Muscle: "Quads"},
			{Name: "Dumbbell Lateral Lunge", HowTo: "Langkah lebar ke samping, squat satu kaki, kembali ke tengah", Reps: "3x10/side", Muscle: "Inner Thighs"},
			{Name: "Dumbbell Calf Raise", HowTo: "Berdiri di tepi, angkat tumit naik tahan, turunkan perlahan", Reps: "4x12", Muscle: "Calves"},
			{Name: "Resistance Band Glute Bridge", HowTo: "Berbaring, band di pinggul, dorong naik tahan 2 detik", Reps: "3x15", Muscle: "Glutes"},
			{Name: "Jump Squat", HowTo: "Squat lalu lompat setinggi mungkin, mendarat lembut", Reps: "3x8", Muscle: "Quads/Glutes"},
		},
	},
	"full_body": {
		{
			{Name: "Dumbbell Thruster", HowTo: "Squat dengan dumbbell di bahu, dorong berdiri sambil press ke atas", Reps: "3x10", Muscle: "Full Body"},
			{Name: "Pull-Up", HowTo: "Pegang bar, tarik tubuh naik hingga dagu melewati bar", Reps: "3x8", Muscle: "Back"},
			{Name: "Dumbbell Renegade Row", HowTo: "Posisi plank, tarik dumbbell ke samping perut bergantian", Reps: "3x10", Muscle: "Back/Core"},
			{Name: "Dumbbell Goblet Squat", HowTo: "Dumbbell di dada, squat hingga paha sejajar lantai", Reps: "3x12", Muscle: "Quads"},
			{Name: "Dumbbell Shoulder Press", HowTo: "Duduk, dorong dumbbell dari bahu ke atas", Reps: "3x10", Muscle: "Shoulders"},
			{Name: "Dumbbell Romanian Deadlift", HowTo: "Pinggul ke belakang, dumbbell turun sejajar kaki, berdiri kembali", Reps: "3x10", Muscle: "Hamstrings"},
			{Name: "Resistance Band Bicep Curl", HowTo: "Injak band, curl pegangan ke bahu", Reps: "3x12", Muscle: "Biceps"},
			{Name: "Plank", HowTo: "Siku di lantai, tubuh lurus, tahan posisi", Reps: "3x45d", Muscle: "Core"},
		},
		{
			{Name: "Dumbbell Deadlift", HowTo: "Dumbbell di depan paha, pinggul dorong belakang lalu kembali berdiri", Reps: "3x10", Muscle: "Posterior Chain"},
			{Name: "Push-Up", HowTo: "Tangan selebar bahu, turunkan dada ke lantai, dorong kembali", Reps: "3x15", Muscle: "Chest"},
			{Name: "Dumbbell Row", HowTo: "Miring tubuh, tarik dumbbell ke samping perut", Reps: "3x12", Muscle: "Back"},
			{Name: "Dumbbell Lunge", HowTo: "Langkah maju, squat, dorong kembali", Reps: "3x10/leg", Muscle: "Legs"},
			{Name: "Dumbbell Shoulder Press", HowTo: "Dorong dumbbell dari bahu ke atas secara vertikal", Reps: "3x10", Muscle: "Shoulders"},
			{Name: "Resistance Band Glute Bridge", HowTo: "Band di pinggul, dorong naik tahan 2 detik", Reps: "3x12", Muscle: "Glutes"},
			{Name: "Dumbbell Bicep Curl", HowTo: "Curl dumbbell dari paha ke bahu", Reps: "3x12", Muscle: "Biceps"},
			{Name: "Bicycle Crunch", HowTo: "Berbaring, siku bertemu lutut bergantian", Reps: "3x20", Muscle: "Core"},
		},
		{
			{Name: "Dumbbell Clean and Press", HowTo: "Angkat dumbbell dari lantai ke bahu, lalu press ke atas", Reps: "3x8", Muscle: "Full Body"},
			{Name: "Chin-Up", HowTo: "Pull-up telapak menghadap ke diri, tarik naik", Reps: "3x6", Muscle: "Back/Biceps"},
			{Name: "Dumbbell Bench Press", HowTo: "Berbaring, dorong dumbbell dari samping dada ke atas", Reps: "3x10", Muscle: "Chest"},
			{Name: "Dumbbell Bulgarian Split Squat", HowTo: "Kaki belakang di bangku, squat dengan kaki depan", Reps: "3x8/leg", Muscle: "Quads"},
			{Name: "Dumbbell Lateral Raise", HowTo: "Angkat dumbbell ke samping hingga sejajar bahu", Reps: "3x12", Muscle: "Shoulders"},
			{Name: "Dumbbell RDL", HowTo: "Kaki sedikit tekuk, dorong pinggul ke belakang, berdiri kembali", Reps: "3x10", Muscle: "Hamstrings"},
			{Name: "Resistance Band Tricep Pushdown", HowTo: "Kaitkan band di atas, dorong pegangan ke bawah hingga lengan lurus", Reps: "3x12", Muscle: "Triceps"},
			{Name: "Mountain Climber", HowTo: "Posisi plank, tarik lutut ke dada bergantian cepat", Reps: "3x20", Muscle: "Core/Cardio"},
		},
		{
			{Name: "Dumbbell Squat to Press", HowTo: "Squat lalu dorong dumbbell ke atas saat berdiri", Reps: "3x10", Muscle: "Full Body"},
			{Name: "Resistance Band Row", HowTo: "Kaitkan band, tarik pegangan ke arah perut", Reps: "3x12", Muscle: "Back"},
			{Name: "Diamond Push-Up", HowTo: "Push-up tangan rapat diamond, fokus tricep", Reps: "3x10", Muscle: "Triceps"},
			{Name: "Dumbbell Walking Lunge", HowTo: "Langkah maju bergantian, dumbbell di samping tubuh", Reps: "3x10/leg", Muscle: "Legs"},
			{Name: "Dumbbell Arnold Press", HowTo: "Curl lalu rotasi dan press ke atas", Reps: "3x10", Muscle: "Shoulders"},
			{Name: "Dumbbell Hammer Curl", HowTo: "Tangan netral, curl ke arah bahu", Reps: "3x10", Muscle: "Biceps"},
			{Name: "Resistance Band Hip Thrust", HowTo: "Band di pinggul, punggung di bangku, dorong naik", Reps: "3x12", Muscle: "Glutes"},
			{Name: "Russian Twist", HowTo: "Duduk miring, kaki terangkat, putar tubuh kiri kanan", Reps: "3x20", Muscle: "Core"},
		},
		{
			{Name: "Dumbbell Thruster", HowTo: "Front squat dengan dumbbell, dorong ke atas saat berdiri", Reps: "4x8", Muscle: "Full Body"},
			{Name: "Pull-Up", HowTo: "Tarik tubuh naik hingga dagu melewati bar", Reps: "3x6", Muscle: "Back"},
			{Name: "Dumbbell Floor Press", HowTo: "Berbaring di lantai, dorong dumbbell ke atas hingga lengan lurus", Reps: "3x10", Muscle: "Chest"},
			{Name: "Dumbbell Step-Up", HowTo: "Naik bangku dengan dumbbell, dorong dengan kaki depan", Reps: "3x10/leg", Muscle: "Legs"},
			{Name: "Dumbbell Upright Row", HowTo: "Tarik dumbbell naik sepanjang tubuh hingga sejajar dada", Reps: "3x10", Muscle: "Shoulders"},
			{Name: "Resistance Band Curl", HowTo: "Injak band, curl ke bahu tahan 1 detik", Reps: "3x12", Muscle: "Biceps"},
			{Name: "Dumbbell Calf Raise", HowTo: "Berdiri, angkat tumit naik turunkan perlahan", Reps: "3x15", Muscle: "Calves"},
			{Name: "Dead Bug", HowTo: "Berbaring telentang, lengan ke atas, turunkan lengan dan kaki bergantian", Reps: "3x10/side", Muscle: "Core"},
		},
	},
}

var WarmUpPool = map[string][]Exercise{
	"push": {
		{Name: "Arm Circles", HowTo: "Berdiri, putar lengan ke depan 20x lalu ke belakang 20x, mulai dari kecil lalu perbesar", Reps: "2x20", Muscle: "Shoulders"},
		{Name: "Band Pull-Apart", HowTo: "Pegang resistance band di depan dada, tarik ke samping hingga band menyentuh dada, kembali perlahan", Reps: "2x15", Muscle: "Rear Delts"},
		{Name: "Push-Up (Lambat)", HowTo: "Push-up dengan tempo lambat: 3 detik turun, 1 detik tahan bawah, 1 detik naik", Reps: "1x10", Muscle: "Chest"},
		{Name: "Dumbbell Front Raise (Ringan)", HowTo: "Pakai dumbbell ringan, angkat ke depan hingga sejajar bahu secara bergantian", Reps: "1x10", Muscle: "Front Delts"},
	},
	"pull": {
		{Name: "Arm Circles", HowTo: "Berdiri, putar lengan ke depan 20x lalu ke belakang 20x", Reps: "2x20", Muscle: "Shoulders"},
		{Name: "Band Pull-Apart", HowTo: "Pegang resistance band di depan dada, tarik ke samping hingga menyentuh dada", Reps: "2x15", Muscle: "Rear Delts"},
		{Name: "Dead Hang", HowTo: "Gantung di pull-up bar, tahan 15-30 detik, rilexas bahu", Reps: "2x15d", Muscle: "Back/Grip"},
		{Name: "Resistance Band Curl (Ringan)", HowTo: "Injak band ringan, curl pegangan ke bahu perlahan, fokus squeeze", Reps: "1x12", Muscle: "Biceps"},
	},
	"legs": {
		{Name: "Leg Swing", HowTo: "Berdiri satu kaki, ayunkan kaki lain ke depan-belakang 15x, ganti", Reps: "1x15/kaki", Muscle: "Hip Flexors"},
		{Name: "Bodyweight Squat", HowTo: "Squat tanpa beban, perlahan dan dalam, fokus mobilitas pinggul", Reps: "1x15", Muscle: "Quads/Hips"},
		{Name: "Walking Lunges", HowTo: "Langkah maju lutut turun, tanpa beban, fokus stabilitas", Reps: "1x10/kaki", Muscle: "Quads/Glutes"},
		{Name: "Calf Stretch", HowTo: "Berdiri di tepi, turunkan tumit ke bawah tahan 15 detik", Reps: "2x15d", Muscle: "Calves"},
	},
	"full_body": {
		{Name: "Jumping Jack", HowTo: "Lompat sambil buka kaki lebar dan tepuk tangan di atas", Reps: "1x30", Muscle: "Full Body"},
		{Name: "Arm Circles", HowTo: "Putar lengan ke depan 15x lalu ke belakang 15x", Reps: "1x15", Muscle: "Shoulders"},
		{Name: "Bodyweight Squat", HowTo: "Squat tanpa beban, perlahan, fokus form", Reps: "1x12", Muscle: "Quads"},
		{Name: "Band Pull-Apart", HowTo: "Pegang band di depan dada, tarik ke samping", Reps: "1x12", Muscle: "Upper Back"},
	},
}

var CoolDownPool = map[string][]Exercise{
	"push": {
		{Name: "Chest Door Stretch", HowTo: "Tangan di kosen pintu, lengan bengkok 90°, putar tubuh ke arah berlawanan tahan 20 detik", Reps: "2x20d/sisi", Muscle: "Chest"},
		{Name: "Overhead Tricep Stretch", HowTo: "Angkat satu tangan ke atas, tekuk siku letakkan tangan di belakang kepala, tarik siku dengan tangan lain tahan 20 detik", Reps: "2x20d/sisi", Muscle: "Triceps"},
		{Name: "Cross Body Shoulder Stretch", HowTo: "Tarik satu lengan menyilang ke dada, tahan 20 detik, ganti", Reps: "2x20d/sisi", Muscle: "Shoulders"},
		{Name: "Child's Pose", HowTo: "Berlutut, duduk di tumit, tangan merembes ke depan, dada ke lantai, tarik napas dalam 5x", Reps: "1x30d", Muscle: "Back/Shoulders"},
	},
	"pull": {
		{Name: "Lat Stretch", HowTo: "Satu tangan di atas pegang pergelangan tangan dengan tangan lain, tarik ke samping tahan 20 detik, ganti", Reps: "2x20d/sisi", Muscle: "Lats"},
		{Name: "Bicep Wall Stretch", HowTo: "Tangan di dinding telapak menghadap ke bawah, putar tubuh perlahan tahan 20 detik", Reps: "2x20d/sisi", Muscle: "Biceps"},
		{Name: "Child's Pose", HowTo: "Berlutut, duduk di tumit, tangan merembes ke depan, dada ke lantai, napas dalam", Reps: "1x30d", Muscle: "Back"},
		{Name: "Wrist Flexor Stretch", HowTo: "Luruskan lengan ke depan, tarik jari ke bawah dengan tangan lain tahan 15 detik", Reps: "2x15d/sisi", Muscle: "Forearms"},
	},
	"legs": {
		{Name: "Standing Quad Stretch", HowTo: "Berdiri satu kaki, tarik kaki ke belakang pegang pergelangan, dorong pinggul ke depan tahan 20 detik", Reps: "2x20d/kaki", Muscle: "Quads"},
		{Name: "Seated Hamstring Stretch", HowTo: "Duduk kaki lurus, raih ujung kaki perlahan, tahan 20 detik", Reps: "2x20d", Muscle: "Hamstrings"},
		{Name: "Pigeon Stretch", HowTo: "Satu kaki tekuk di depan, kaki lain lurus ke belakang, miring tubuh ke depan tahan 20 detik", Reps: "2x20d/sisi", Muscle: "Glutes/Hips"},
		{Name: "Calf Wall Stretch", HowTo: "Tangan di dinding, satu kaki depan tekuk, kaki belakang lurus tumit di lantai, tahan 20 detik", Reps: "2x20d/kaki", Muscle: "Calves"},
	},
	"full_body": {
		{Name: "Standing Forward Fold", HowTo: "Berdiri, tekuk tubuh ke depan, raih kaki/lantai, rilexas leher, tahan 20 detik", Reps: "1x30d", Muscle: "Hamstrings/Back"},
		{Name: "Chest Door Stretch", HowTo: "Tangan di kosen, putar tubuh ke arah berlawanan tahan 20 detik", Reps: "2x20d/sisi", Muscle: "Chest"},
		{Name: "Standing Quad Stretch", HowTo: "Tarik kaki ke belakang pegang pergelangan, tahan 20 detik per sisi", Reps: "1x20d/kaki", Muscle: "Quads"},
		{Name: "Child's Pose", HowTo: "Berlutut, tangan merembes ke depan, dada ke lantai, tarik napas dalam 5x", Reps: "1x30d", Muscle: "Full Body"},
	},
}

func GetWarmUp(workoutType string) []Exercise {
	return WarmUpPool[workoutType]
}

func GetCoolDown(workoutType string) []Exercise {
	return CoolDownPool[workoutType]
}

var DayWorkoutMap = map[int]string{
	1: "push",
	2: "legs",
	4: "pull",
	5: "full_body",
}

var WorkoutTypeNames = map[string]string{
	"push":      "Push (Dada, Bahu, Tricep)",
	"pull":      "Pull (Punggung, Bicep)",
	"legs":      "Legs (Kaki, Glute)",
	"full_body": "Full Body (Seluruh Tubuh)",
}

var DayNames = map[int]string{
	1: "Senin",
	2: "Selasa",
	3: "Rabu",
	4: "Kamis",
	5: "Jumat",
	6: "Sabtu",
	7: "Minggu",
}

func GetWorkoutForDay(dayOfWeek int) []Exercise {
	workoutType, ok := DayWorkoutMap[dayOfWeek]
	if !ok {
		return nil
	}

	pool, ok := WorkoutPool[workoutType]
	if !ok || len(pool) == 0 {
		return nil
	}

	_, week := time.Now().ISOWeek()
	index := (week - 1) % len(pool)

	return pool[index]
}

func GetWorkoutType(dayOfWeek int) string {
	return DayWorkoutMap[dayOfWeek]
}
