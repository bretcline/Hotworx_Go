package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bretcline/Hotworx_Go/hotworxData"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

	//DataMigration(user, password)

	data, success := GetDailyLeaderboard(hotworx_api, true)
	if success {
		fmt.Println(data)
		// TODO: Do something about the error
		WriteToDB(user, password, "Leaderboard_RAW", data)

		y := make([]interface{}, len(data.Leaderboard))
		for i, v := range data.Leaderboard {
			v.Date = data.Date
			y[i] = v
		}
		WriteMultipleToDB(user, password, "Leaderboard", y)
	}

	// summary, success := GetDailySummary(hotworx_api, true)
	// if success {
	// 	fmt.Println(data)
	// 	// TODO: Do something about the error
	// 	WriteToDB(user, password, "DailySummary_RAW", summary)
	// }

	fmt.Println("done")
}

func DataMigration(user string, password string) {

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

	rcol := client.Database("Hotworx").Collection("Leaderboard_RAW")

	cursor, err := rcol.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var results []hotworxData.LeaderboardResults
	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}

	for _, leader := range results {
		y := make([]interface{}, len(leader.Leaderboard))
		for i, v := range leader.Leaderboard {
			v.Date = leader.Date
			y[i] = v
		}
		WriteMultipleToDB(user, password, "Leaderboard", y)
	}

	client.Disconnect(ctx)
}

func WriteToDB[V hotworxData.LeaderboardResults | hotworxData.DailyResults | []hotworxData.Leaderboard](user string, password string, collection string, data V) {
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

func WriteMultipleToDB(user string, password string, collection string, data []interface{}) {
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

	results, e := col.InsertMany(ctx, data)

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
	req.Header.Add("Content-Type", "application/json,text/html,application/xhtml+xml,application/xml")

	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, true
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		body = nil
	} else {
		rc = true
	}
	httpClient.CloseIdleConnections()
	return body, rc
}

func GetDailySummary(hotworx_api string, writeFile bool) (hotworxData.DailyResults, bool) {
	rc := false
	const layout = "Jan 02,2006"
	t := time.Now()

	params := url.Values{}
	params.Add("action", "get_summary")
	params.Add("user_id", "841f1f22368a4b5d39b4838016ea5a51")
	params.Add("date", t.Format(layout))

	url := fmt.Sprintf("%s?%s", hotworx_api, params.Encode())
	//url := fmt.Sprintf("%s?action=get_summary&user_id=841f1f22368a4b5d39b4838016ea5a51", hotworx_api)

	body, rc := CallAPI(url, "841f1f22368a4b5d39b4838016ea5a51")

	var data hotworxData.DailyResults

	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("failed to unmarshal:", err)
	} else {
		data.Date = time.Now()
		if writeFile {
			body, _ := json.Marshal(data)
			jsonString, _ := PrettyString(string(body))
			fileName := GetFilenameDate("Daily")
			byteData := []byte(jsonString)
			os.WriteFile(fileName, byteData, 0644)
		}
	}
	return data, rc
}

func GetDailyLeaderboard(hotworx_api string, writeFile bool) (hotworxData.LeaderboardResults, bool) {
	rc := false

	params := url.Values{}
	params.Add("action", "get_user_leaderboard_local")
	params.Add("user_id", "841f1f22368a4b5d39b4838016ea5a51")

	url := fmt.Sprintf("%s?%s", hotworx_api, params.Encode())

	body, rc := CallAPI(url, "841f1f22368a4b5d39b4838016ea5a51")

	var data hotworxData.LeaderboardResults

	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("failed to unmarshal:", err)
	} else {
		data.Date = time.Now()
		if writeFile {
			body, _ := json.Marshal(data)
			jsonString, _ := PrettyString(string(body))
			fileName := GetFilenameDate("data")
			byteData := []byte(jsonString)
			os.WriteFile(fileName, byteData, 0644)
		}
	}
	return data, rc
}

// func GetData[V hotworxData.LeaderboardResults | hotworxData.DailyResults](writeFile bool, body []byte, data V) {
// 	if err := json.Unmarshal(body, &data); err != nil {
// 		fmt.Println("failed to unmarshal:", err)
// 	} else {
// 		data.Date = time.Now()
// 		if writeFile {
// 			body, _ := json.Marshal(data)
// 			jsonString, _ := PrettyString(string(body))
// 			fileName := GetFilenameDate("data")
// 			byteData := []byte(jsonString)
// 			os.WriteFile(fileName, byteData, 0644)
// 		}
// 	}
// }

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
