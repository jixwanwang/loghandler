loghandler
==========

The default Go HTTP handler doesn't do any logging. This is a handler middleware package that logs to an [`io.Writer`](http://golang.org/pkg/io/#Writer) and [StatsD](https://github.com/etsy/statsd) (optionally).

    go get github.com/Hebo/loghandler
    
loghandler is based on Gorilla's [handler code](http://www.gorillatoolkit.org/pkg/handlers), and adds support for StatsD and tracking request duration.

### Example Output

```
199.9.252.10 - - [03/Feb/2014:16:32:41 -0800] "GET /validate HTTP/1.1" 200 239 (2424μs)
199.9.250.239 - - [03/Feb/2014:16:32:41 -0800] "GET /validate HTTP/1.1" 200 244 (2490μs)
199.9.252.10 - - [03/Feb/2014:16:32:41 -0800] "GET /validate HTTP/1.1" 200 235 (3024μs)
```


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

