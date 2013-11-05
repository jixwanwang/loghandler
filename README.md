loghandler
==========

The default Go HTTP handler doesn't do any logging. This is a handler middleware package that logs to an io.Writer and StatsD (optionally).

    go get github.com/Hebo/loghandler
    
loghandler is based on Gorilla's [handler code](http://www.gorillatoolkit.org/pkg/handlers), and adds support for StatsD and tracking request duration.

## Getting Started

```
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
```

To add stats support, simply change the `nil` in `NewLoggingHandler` to something that implements `StatsLogger`.

So Easy!
