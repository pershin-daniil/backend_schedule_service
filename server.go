package main

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func timeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/timeNow" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	stateTime = append(stateTime, time.Now())
	_, err := fmt.Fprintf(w, "%s\n", stateTime[len(stateTime)-1])
	if err != nil {
		log.Panic(err)
	}
}
func timeHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/timeHistory" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	_, err := fmt.Fprintf(w, "%v\n", stateTime)
	if err != nil {
		log.Panic(err)
	}
}

func resetTimeHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/resetTimeHistory" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	_, err := fmt.Fprintf(w, "%v\n", stateTime)
	if err != nil {
		log.Panic(err)
	}
	stateTime = make([]time.Time, 0)
}

var stateTime = make([]time.Time, 0)

func main() {
	http.HandleFunc("/timeNow", timeHandler)
	http.HandleFunc("/timeHistory", timeHistoryHandler)
	http.HandleFunc("/resetTimeHistory", resetTimeHistoryHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
