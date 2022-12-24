package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
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

	_, err := fmt.Fprintf(w, "%s\n", time.Now())
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	http.HandleFunc("/timeNow", timeHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
