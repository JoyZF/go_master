package main

import (
	"fmt"
	"log"
	"net/http"
	"syscall"
	"time"
)

func main() {
	fmt.Println(syscall.Getrlimit(1, &syscall.Rlimit{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		for i := 0; i < 100; i++ {
			fmt.Fprintf(w, "data: Message %d\n\n", i)
			w.(http.Flusher).Flush()
			time.Sleep(1 * time.Second)
		}
	})

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
