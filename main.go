package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Workiva/go-datastructures/queue"
)

var (
	queueOfMessagesFromAgent = queue.New(10)
	queueOfDataToDB          *queue.Queue
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
	for {
		data, err := queueOfMessagesFromAgent.Get(1)
		//TODO: обработка ошибки
		if err != nil {
		}
		scanner := bufio.NewScanner(strings.NewReader(string(data[0].([]byte))))
		for scanner.Scan() {
			logInfo := getLogInfoFromString(scanner.Text())
			fmt.Println(logInfo)
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

func writeToDB() {}

func initDB() {}
