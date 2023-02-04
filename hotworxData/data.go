package hotworxData

import "time"

type LeaderboardResults struct {
	Status      bool          `bson:"status"`
	Message1    string        `bson:"message1"`
	Leaderboard []Leaderboard `bson:"leaderboard"`
	Date        time.Time     `bson:"Date,omitempty"`
}

type Leaderboard struct {
	User_Id            string    `bson:"user_id"`
	TotalCaloriesBurnt string    `bson:"TotalCaloriesBurnt"`
	Reward             string    `bson:"reward"`
	Username           string    `bson:"username"`
	Selft_Entry        string    `bson:"selft_entry"`
	Date               time.Time `bson:"Date,omitempty"`
}

type DailyResults struct {
	Status           bool               `bson:"status"`
	Message          string             `bson:"message"`
	Summary          []Summary          `bson:"summary"`
	ClassesCompleted []ClassesCompleted `bson:"classes_completed,omitempty"`
	Date             time.Time          `bson:"Date,omitempty"`
}

type Summary struct {
	IsometricCalories string `bson:"isometric_calories"`
	HIITCalories      string `bson:"hiit_calories"`
	AfterBurn         string `bson:"after_burn"`
}

type ClassesCompleted struct {
	Type         string `bson:"type"`
	BurnCalories string `bson:"burn_calories"`
}
