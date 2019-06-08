package main

import (
	"bufio"
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/alecthomas/template"
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
	deviceID string
}

type ReportOptions struct {
	Levels    []string
	Types     []string
	AgentsIDs []string
}

func main() {
	connectToDB()
	go processData()
	http.HandleFunc("/logs", logsHandler)
	http.HandleFunc("/report", reportHandler)
	panic(http.ListenAndServe(":8080", nil))
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("template/report.html")
	if err != nil {
		panic(err)
	}
	repOpt := ReportOptions{
		getLogsLevels(),
		getDevicesNames(),
		getAgentsIDs(),
	}
	tmpl.Execute(w, repOpt)
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	queueOfMessagesFromAgent.Put(body)
	w.Write([]byte("succes"))
}

//TODO: отчет по логам
func processData() {
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
	_, err := strconv.Atoi(data[len(data)-1])
	var message string
	var deviceID string
	if err != nil {
		message = strings.Join(data[5:], " ")
		deviceID = ""
	} else {
		message = strings.Join(data[5:len(data)-2], " ")
		deviceID = data[len(data)-1]
	}
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
		deviceID,
	}
	return logInfo
}

func writeToDBlogInfo(db *sql.DB, logInfo logInfo) {
	_, err := db.Exec("INSERT INTO log(agent_id, time, function, level, id, message, device_id) VALUES ($1, $2, $3, $4, $5, $6, $7)", logInfo.agentID, logInfo.time, logInfo.function, logInfo.level, logInfo.id, logInfo.message, logInfo.deviceID)
	//TODO: обработать ошибку
	if err != nil {
		panic(err)
	}
}

func connectToDB() {
	connStr := "user=postgres password=postgres dbname=carwashing sslmode=disable host=localhost port=5432"
	dataBase, err := sql.Open("postgres", connStr)
	//TODO: обработать ошибку
	if err != nil {
		log.Fatal(err)
	}
	db = dataBase

}

func getAgentsIDs() []string {
	rows, err := db.Query("SELECT DISTINCT agent_id FROM log")
	if err != nil {
		panic(err)
	}
	result := make([]string, 0)
	for rows.Next() {
		tmp := ""
		err = rows.Scan(&tmp)
		result = append(result, tmp)
	}
	return result
}

func getDevicesNames() []string {
	rows, err := db.Query("SELECT DISTINCT device.name FROM log INNER JOIN device ON log.device_id=device.id")
	if err != nil {
		panic(err)
	}
	result := make([]string, 0)
	for rows.Next() {
		tmp := ""
		err = rows.Scan(&tmp)
		result = append(result, tmp)
	}
	return result
}

func getLogsLevels() []string {
	rows, err := db.Query("SELECT DISTINCT level FROM log")
	if err != nil {
		panic(err)
	}
	result := make([]string, 0)
	for rows.Next() {
		tmp := ""
		err = rows.Scan(&tmp)
		result = append(result, tmp)
	}
	return result
}
