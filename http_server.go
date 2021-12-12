package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "log"
    "sync"
    "sync/atomic"
    "os/signal"
    "strconv"
    "strings"
    "time"
    model "github.com/rikesh-chouhan/go-http-server/model"
)

var map_of_texts = make(map[int64]model.TimedData)
var map_of_hashed = make(map[int64]string)
var hashAverage int64
var shutdown bool
var requests int64

// increments the number of requests and returns the new value
func IncRequests() int64 {
    return atomic.AddInt64(&requests, 1)
}

// returns the current value
func GetRequests() int64 {
    return atomic.LoadInt64(&requests)
}

// increments the number of requests and returns the new value
func IncAverage(currentReqs int64, currentAvg int64, newValue int64) int64 {
    return atomic.AddInt64(&hashAverage, ((currentReqs-1)*currentAvg+newValue)/currentReqs)
}

// returns the current value
func GetAverage() int64 {
    return atomic.LoadInt64(&requests)
}

func main() {
    http.HandleFunc("/shutdown", ShutdownHandler)
    http.HandleFunc("/stats", StatsHandler)
    http.HandleFunc("/hash", HashHandler)
    // used for debugging only
    http.HandleFunc("/entries", EntriesHandler)
    // Creating a waiting group that waits until the graceful shutdown procedure is done
    var wg sync.WaitGroup
    wg.Add(1)

    server := &http.Server{Addr:":8080", Handler: nil}
    // This goroutine is running in parallels to the main one
    go func() {
        // creating a channel to listen for signals, like SIGINT
        stop := make(chan os.Signal, 1)
        // subscribing to interruption signals
        signal.Notify(stop, os.Interrupt)
        // this blocks until the signal is received
        <-stop
        // initiating the shutdown
        err := server.Shutdown(context.Background())
        // can't do much here except for logging any errors
        if err != nil {
            log.Printf("error during shutdown: %v\n", err)
        }
        // notifying the main goroutine that we are done
        wg.Done()
    }()

    log.Println("Server started at port 8080")
    err := server.ListenAndServe()
    if err == http.ErrServerClosed { // graceful shutdown
        log.Println("commencing server shutdown...")
        wg.Wait()
        log.Println("server was gracefully shut down.")
    } else if err != nil {
        log.Printf("server error: %v\n", err)
    }
}

/*
Calculate the hash asynchronously when it has been at least 4 seconds after the request
was made.
*/
func HashCalculator() {
}

/*
 Stats handler returns stats for an endpoint.
 */
func StatsHandler(w http.ResponseWriter, r *http.Request) {
    if !shutdown {
        w.Header().Set("Content-Type", "application/json")
        switch r.Method {
        case "GET":
            log.Println("Writing stats")
            fmt.Fprintf(w, `{"total": %d "average": %d}` + "\n", GetRequests(), GetAverage())
        default:
            w.WriteHeader(http.StatusNotFound)
            w.Write([]byte(`{"error": "only GET supported"}`))
        }
    }
}

// set shutdown flag to true to stop handling new hash requests.
func ShutdownHandler(w http.ResponseWriter, r *http.Request) {
    if !shutdown {
        shutdown = true
        log.Println("Shutting down")
        fmt.Fprintf(w, "Shutting down server\n")
    }
}

// used for debugging
func EntriesHandler(w http.ResponseWriter, r *http.Request) {
    if !shutdown {
        w.Header().Set("Content-Type", "application/json")
        jmap, err := json.Marshal(map_of_texts)
        if err == nil {
            fmt.Fprintf(w, "%s\n", string(jmap))
        }
    }
}

// Process hash generation service
func HashHandler(w http.ResponseWriter, r *http.Request)  {
    if !shutdown {
        reqNum := IncRequests()
        currentAvg := GetAverage()
        start := time.Now()
        w.Header().Set("Content-Type", "text/plain")
        switch r.Method {
        case "GET":
            numPath := r.URL.Path[1:]
            fmt.Fprintf(w, "Number provided %s!\n", numPath)
            if len(strings.TrimSpace(numPath)) > 0 {
                numExtracted, err := strconv.ParseInt(numPath, 10, 64)
                if err == nil {
                    fmt.Fprintf(w, "%d provided\n", numExtracted)
                }
            }
        case "POST":
            if err := r.ParseForm(); err != nil {
                fmt.Fprintf(w, "ParseForm() err: %v", err)
                return
            }

            pwParam := r.FormValue("password")
            if len(strings.TrimSpace(pwParam)) < 1 {
                fmt.Fprintf(w, "Form variable password is empty\n")
                return
            }
            map_of_texts[reqNum] = model.NewTimedData(reqNum, pwParam)
            fmt.Fprintf(w, "%d\n", reqNum)

        default:
            fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
        }
        executeTime := time.Now().Sub(start).Microseconds()
        IncAverage( reqNum, currentAvg, executeTime)
    }
}

