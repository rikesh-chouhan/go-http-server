package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "time"
    "log"
    "sync"
    "sync/atomic"
)

var map_of_texts = sync.Map
var map_of_hashed = sync.Map
var map_of_stats = sync.Map
var shutdown bool
var requests int64

// increments the number of requests and returns the new value
func incRequests() int64 {
    return atomic.AddInt64(&requests, 1)
}

// returns the current value
func getRequests() int64 {
    return atomic.LoadInt64(&requests)
}

func main() {
    http.HandleFunc("/shutdown", ShutdownHandler)
    http.HandleFunc("/stats", StatsHandler)
    // Creating a waiting group that waits until the graceful shutdown procedure is done
    var wg sync.WaitGroup
    wg.Add(1)

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

func SendPostAsync(url string, body []byte, rc chan *http.Response) error {
    response, err := http.Post(url, "application/json", bytes.NewReader(body))
    if err == nil {
        rc <- response
    }

    return err
}

// Handler function that records method execution time.
func timer(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        startTime := time.Now()
        h.ServeHTTP(w, r)
        duration := time.Now().Sub(startTime).Microseconds()
    })
}

/*
 Stats handler returns stats for an endpoint.
 */
func StatsHandler(w http.ResponseWriter, r *http.Request) error {
    if shutdown {
        ServiceStopping(w, r)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    switch r.Method {
    case "GET":
        log.Println("Writing stats")
        jmap, err := json.Marshal(map_of_stats)
        if err != nil {
            fmt.Printf("Error: %s", err.Error())
        } else {
            w.Write([]byte(string(jmap)))
            fmt.Println(string(jmap))
        }

    default:
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte(`{"message": "not found"}`))
    }
}

// set shutdown flag to true to stop handling new hash requests.
func ShutdownHandler(w http.ResponseWriter, r *http.Request) error {
    shutdown = true
    log.Println("Shutting down")
    ServiceStopping(w, r)
}

func ServiceStopping(w http.ResponseWriter, r *http.Request) error {
    fmt.Fprintf(w, "Shutting down server\n")
}