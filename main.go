package main

import (
	"bufio"
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Workiva/go-datastructures/queue"
	_ "github.com/lib/pq"
)

var (
	queueOfMessagesFromAgent = queue.New(10)
	queueOfDataToDB          *queue.Queue
	db                       *sql.DB
)

type logInfo struct {
	agentID  string
	time     time.Time
	function string
	level    string
	id       string
	message  string
}

func main() {
	go processData()
	http.HandleFunc("/logs", logsHandler)
	panic(http.ListenAndServe(":8080", nil))
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	queueOfMessagesFromAgent.Put(body)
	w.Write([]byte("succes"))
}

func processData() {
	db := connectToDB()
	for {
		data, err := queueOfMessagesFromAgent.Get(1)
		//TODO: обработка ошибки
		if err != nil {
		}
		scanner := bufio.NewScanner(strings.NewReader(string(data[0].([]byte))))
		for scanner.Scan() {
			logInfo := getLogInfoFromString(scanner.Text())
			writeToDBlogInfo(db, logInfo)
		}
	}
}

func getLogInfoFromString(log string) logInfo {
	data := strings.Split(log, " ")
	message := strings.Join(data[5:], " ")
	timeString := strings.Join(data[1:3], " ")
	t, err := time.Parse("2006-01-02 15:04:05", timeString)
	//TODO: обработать ошибку
	if err != nil {

	}
	logInfo := logInfo{
		data[0],
		t,
		data[3],
		data[4],
		data[5],
		message,
	}
	return logInfo
}

func writeToDBlogInfo(db *sql.DB, logInfo logInfo) {
	_, err := db.Exec("INSERT INTO log(agent_id, time, function, level, id, message) VALUES ($1, $2, $3, $4, $5, $6)", logInfo.agentID, logInfo.time, logInfo.function, logInfo.level, logInfo.id, logInfo.message)
	//TODO: обработать ошибку
	if err != nil {
		panic(err)
	}
}

func connectToDB() *sql.DB {
	connStr := "user=postgres password=postgres dbname=carwashing sslmode=disable host=localhost port=5432"
	db, err := sql.Open("postgres", connStr)
	//TODO: обработать ошибку
	if err != nil {
		log.Fatal(err)
	}
	return db
}
