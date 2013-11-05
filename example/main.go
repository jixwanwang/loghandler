package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Hebo/loghandler"
)

func main() {
	http.HandleFunc("/dowork", func(w http.ResponseWriter, r *http.Request) {
		// Set statsd stat name
		loghandler.SetStat(w, "do.work")
		time.Sleep(300 * time.Millisecond)
		fmt.Fprintf(w, "Hello!")
	})

	// Wrap default handler and serve that instead
	handler := loghandler.NewLoggingHandler(os.Stdout, nil, http.DefaultServeMux)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Could not start server: %s", err.Error())
	}
}
