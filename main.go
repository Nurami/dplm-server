package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/logs", logsHandler)
	panic(http.ListenAndServe(":8080", nil))
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	w.Write([]byte("succes"))
}
