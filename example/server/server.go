package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type stat struct {
	Mutex            sync.Mutex
	T                time.Time
	NumberOfRequests int
}

var S stat

func serveHttp(w http.ResponseWriter, r *http.Request) {
	S.Mutex.Lock()
	if S.NumberOfRequests == 0 {
		S.T = time.Now()
	}

	S.NumberOfRequests = S.NumberOfRequests + 1
	n := S.NumberOfRequests
	t := S.T
	S.Mutex.Unlock()

	fmt.Println("Received request at time :: ", time.Since(t))
	fmt.Println("Number of requests so far :: ", n)
}

func main() {

	http.HandleFunc("/", serveHttp)

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err := s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
