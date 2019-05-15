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
	data, err := queueOfMessagesFromAgent.Get(1)
	//TODO: обработка ошибки
	if err != nil {
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data[0].([]byte))))
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func writeToDB() {}

func initDB() {}
