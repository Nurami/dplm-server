package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Workiva/go-datastructures/queue"
)

var (
	queueOfMessagesFromAgent = queue.New(10)
	queueOfDataToDB          *queue.Queue
)

type logInfo struct {
	agentID  string
	time     string
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
	logInfo := logInfo{
		data[0],
		data[1],
		data[2],
		data[3],
		data[4],
		message,
	}
	return logInfo
}

func writeToDB() {}

func initDB() {}
