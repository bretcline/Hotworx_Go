package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LeaderboardResults struct {
	Status      bool          `bson:"status"`
	Message1    string        `bson:"message1"`
	Leaderboard []Leaderboard `bson:"leaderboard"`
	Date        time.Time     `bson:"Date,omitempty"`
}

type Leaderboard struct {
	User_Id            string `bson:"user_id"`
	TotalCaloriesBurnt string `bson:"TotalCaloriesBurnt"`
	Reward             string `bson:"reward"`
	Username           string `bson:"username"`
	Selft_Entry        string `bson:"selft_entry"`
}

type DailyResults struct {
	Status           bool               `bson:"status"`
	Message          string             `bson:"message"`
	Summary          []Summary          `bson:"summary"`
	ClassesCompleted []ClassesCompleted `bson:"classes_completed"`

	Date time.Time `bson:"Date,omitempty"`
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

func main() {

	envErr := godotenv.Load()
	if envErr != nil {
		fmt.Printf("Error loading credentials: %v", envErr)
	}

	var (
		password    = os.Getenv("MONGO_PASSWORD")
		user        = os.Getenv("MONGO_USER")
		hotworx_api = os.Getenv("HOTWORX_API")
	)

	data, success := GetDailyLeaderboard(hotworx_api, true)
	if success {
		// TODO: Do something about the error
		WriteToDB(user, password, "Leaderboard_RAW", data)
	}

	summary, success := GetDailySummary(hotworx_api, true)
	if success {
		// TODO: Do something about the error
		WriteToDB(user, password, "DailySummary_RAW", summary)
	}

	fmt.Println("done")
}

func WriteToDB[V LeaderboardResults | DailyResults](user string, password string, collection string, data V) {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	connectionString := fmt.Sprintf("mongodb+srv://%s:%s@cluster0.ie1zp2i.mongodb.net/?retryWrites=true&w=majority", user, password)
	clientOptions := options.Client().
		ApplyURI(connectionString).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	col := client.Database("Hotworx").Collection(collection)

	results, e := col.InsertOne(ctx, data)

	if e != nil {

		fmt.Println("Results All: ", e)
	} else {
		fmt.Println("Results All: ", results)
	}
	client.Disconnect(ctx)
}

func CallAPI(url string, userId string) ([]byte, bool) {
	rc := false
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("{\"user_id\": \"%s\"}", userId))

	httpClient := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, true
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, true
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		body = nil
	} else {
		rc = true
	}
	httpClient.CloseIdleConnections()
	return body, rc
}

func GetDailySummary(hotworx_api string, writeFile bool) (DailyResults, bool) {
	rc := false
	const layout = "Jan 02,2006"
	t := time.Now()

	url := fmt.Sprintf("%s?action=get_summary&user_id=841f1f22368a4b5d39b4838016ea5a51&date=%s", hotworx_api, t.Format(layout))

	body, rc := CallAPI(url, "841f1f22368a4b5d39b4838016ea5a51")

	var data DailyResults

	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("failed to unmarshal:", err)
	} else {
		data.Date = time.Now()
		if writeFile {
			body, _ := json.Marshal(data)
			jsonString, _ := PrettyString(string(body))
			fileName := GetFilenameDate("Daily")
			byteData := []byte(jsonString)
			err = ioutil.WriteFile(fileName, byteData, 0644)
		}
	}
	return data, rc
}

func GetDailyLeaderboard(hotworx_api string, writeFile bool) (LeaderboardResults, bool) {
	rc := false
	url := fmt.Sprintf("%s?action=get_user_leaderboard_local&user_id=841f1f22368a4b5d39b4838016ea5a51", hotworx_api)

	body, rc := CallAPI(url, "841f1f22368a4b5d39b4838016ea5a51")

	var data LeaderboardResults

	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("failed to unmarshal:", err)
	} else {
		data.Date = time.Now()
		if writeFile {
			body, _ := json.Marshal(data)
			jsonString, _ := PrettyString(string(body))
			fileName := GetFilenameDate("data")
			byteData := []byte(jsonString)
			err = ioutil.WriteFile(fileName, byteData, 0644)
		}
	}
	return data, rc
}

func GetFilenameDate(prefix string) string {
	// Use layout string for time format.
	const layout = "01-02-2006"
	// Place now in the string.
	t := time.Now()
	return prefix + "-" + t.Format(layout) + ".json"
}

func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
