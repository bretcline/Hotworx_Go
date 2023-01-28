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
)

type Leaderboard struct {
	user_id            string
	TotalCaloriesBurnt int
	reward             string
	username           string
	selft_entry        string
}

type LeaderboardResults struct {
	status      bool
	message1    string
	leaderboard []Leaderboard
}

func main() {

	envErr := godotenv.Load()
	if envErr != nil {
		fmt.Printf("Error loading credentials: %v", envErr)
	}

	var (
		password = os.Getenv("MSSQL_DB_PASSWORD")
		user     = os.Getenv("MSSQL_DB_USER")
	)

	// connectionString := fmt.Sprintf("user id=%s;password=%s;port=%s;database=%s", user, password, port, database)

	// sqlObj, connectionError := sql.Open("mssql", connectionString); if connectionError != nil {
	//    fmt.Println(fmt.Errorf("error opening database: %v", connectionError))
	// }

	// data := database.Database{
	//    SqlDb: sqlObj,
	// }

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	connectionString := fmt.Sprintf("mongodb+srv://%s:%s@cluster0.ie1zp2i.mongodb.net/?retryWrites=true&w=majority", user, password)
	clientOptions := options.Client().
		ApplyURI(connectionString). // "mongodb+srv://bretcline:<password>@cluster0.ie1zp2i.mongodb.net/?retryWrites=true&w=majority").
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	coll := client.Database("Hotworx").Collection("")
	col.InsertOne()

	url := "http://appadminnew.hotworx.net/mobileservice/mobileservice_intermittent.php/mobileservice_intermittent.php?action=get_user_leaderboard_local&user_id=841f1f22368a4b5d39b4838016ea5a51"
	method := "POST"

	payload := strings.NewReader(`{"user_id": "61fdbcc2e4bcce9adb60a84350374222"}`)

	httpClient := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	//body, err = json.MarshalIndent(string(body), "", "    ")

	jsonString, err := PrettyString(string(body))

	fmt.Println(jsonString)
	fileName := GetFilenameDate()

	data := []byte(jsonString)

	err = ioutil.WriteFile(fileName, data, 0644)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("done")

}

func GetFilenameDate() string {
	// Use layout string for time format.
	const layout = "01-02-2006"
	// Place now in the string.
	t := time.Now()
	return "data-" + t.Format(layout) + ".json"
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
