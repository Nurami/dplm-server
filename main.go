package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/alecthomas/template"
	"github.com/op/go-logging"
)

var (
	queueOfMessagesFromAgent = queue.New(10)
	queueOfDataToDB          *queue.Queue
	db                       *sql.DB
	log                      = logging.MustGetLogger("logger")
	logsFormat               = logging.MustStringFormatter(`%{time:2006-01-02 15:04:05} %{shortfunc} %{level:s} %{id:d} %{message}`)
	mutex                    = &sync.Mutex{}
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

type resultInfo struct {
	AgentID string `json:"agentID"`
	Level   string `json:"level"`
	Time    string `json:"time"`
	Message string `json:"message"`
	Name    string `json:"name"`
}

type clientRequest struct {
	AgentID    string `json:"agentID"`
	FirstDate  string `json:"firstDate"`
	SecondDate string `json:"secondDate"`
	Level      string `json:"level"`
	DeviceType string `json:"deviceType"`
}

type ReportOptions struct {
	Levels    []string
	Types     []string
	AgentsIDs []string
}

func main() {
	go logToNewFileByPeriod(10)
	time.Sleep(time.Second)

	connectToDB()

	go processData()

	http.HandleFunc("/logs", logsHandler)
	http.HandleFunc("/report", reportHandler)
	http.HandleFunc("/data", dataHandler)
	log.Panic(http.ListenAndServe(":8080", nil))
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("template/report.html")
	if err != nil {
		log.Error(err)
	}
	repOpt := ReportOptions{
		getLogsLevels(),
		getDevicesNames(),
		getAgentsIDs(),
	}
	tmpl.Execute(w, repOpt)
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
	}
	clRe := clientRequest{}
	err = json.Unmarshal(body, &clRe)
	if err != nil {
		log.Error(err)
	}
	w.Write(getResultLogInfos(clRe))
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
	}
	queueOfMessagesFromAgent.Put(body)
	w.Write([]byte("Succes"))
}

func processData() {
	for {
		data, err := queueOfMessagesFromAgent.Get(1)
		if err != nil {
			log.Error(err)
		}
		scanner := bufio.NewScanner(strings.NewReader(string(data[0].([]byte))))
		for scanner.Scan() {
			logInfo := getLogInfoFromString(scanner.Text())
			writeToDBlogInfo(logInfo)
		}
	}
}

func getLogInfoFromString(currentLog string) logInfo {
	data := strings.Split(currentLog, " ")
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
	if err != nil {
		log.Error(err)
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

func logToNewFileByPeriod(period int) {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}
	for {
		logsName := "logs/" + time.Now().Format("2006.01.02-15.04.05") + ".log"
		file, err := os.Create(logsName)
		mutex.Lock()
		log = logging.MustGetLogger(logsName)
		var backend *logging.LogBackend
		if err != nil {
			fmt.Println(time.Now(), " ", err)
			backend = logging.NewLogBackend(os.Stdout, "", 0)
		} else {
			backend = logging.NewLogBackend(file, "", 0)
		}
		backendFormatter := logging.NewBackendFormatter(backend, logsFormat)
		logging.SetBackend(backendFormatter)
		mutex.Unlock()
		time.Sleep(time.Duration(period) * time.Second)
		file.Close()
	}
}
