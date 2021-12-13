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

var map_of_input = make(map[int64]model.TimedData)
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
    http.HandleFunc("/", HashRespHandler)
    // Creating a waiting group that waits until the graceful shutdown procedure is done
    var wg sync.WaitGroup
    wg.Add(1)

    server := &http.Server{Addr:":8080", Handler: nil}
    stopHashTicker := make(chan bool)

    // Graceful shutdown - capture OS signal
    go func() {
        // creating a channel to listen for signals, like SIGINT
        stop := make(chan os.Signal, 1)
        // subscribing to interruption signals
        signal.Notify(stop, os.Interrupt)
        // this blocks until the signal is received
        <-stop
        stopHashTicker <- true
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
    // start the hashcalculator goroutine
    go HashCalculator(&wg, stopHashTicker, 5)

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
func HashCalculator(waitGroup *sync.WaitGroup, stop <-chan bool, timeToCheck float64) {
	ticker := time.NewTicker(500 * time.Millisecond)

    func() {
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
                for key, element := range map_of_input {
                    if time.Since(element.Start).Seconds() >  timeToCheck {
                        map_of_hashed[key] = model.GenerateHash(&element)
                        delete(map_of_input, key)
                    }
                }
			}
		}
    }()

    log.Println("Outside of ticker loop")
    ticker.Stop()
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
        jmap, err := json.Marshal(map_of_input)
        if err == nil {
            fmt.Fprintf(w, "%s\n", string(jmap))
        }
    }
}

func HashRespHandler(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    var id int64

    switch {
    case match(path, "/hash/+", &id):
        log.Printf("Number provided: %v\n", id)
        if x, found := map_of_hashed[id]; found {
            fmt.Fprintf(w, "%s\n", x)
        } else {
            fmt.Fprintf(w, "%d does not have a hash value\n", id)
        }

	default:
		http.NotFound(w, r)
		return
	}
}

// match reports whether path matches the given pattern, which is a
// path with '+' wildcards wherever you want to use a parameter. Path
// parameters are assigned to the pointers in vars (len(vars) must be
// the number of wildcards), which must be of type *string or *int.
func match(path, pattern string, vars ...interface{}) bool {
	for ; pattern != "" && path != ""; pattern = pattern[1:] {
		switch pattern[0] {
		case '+':
			// '+' matches till next slash in path
			slash := strings.IndexByte(path, '/')
			if slash < 0 {
				slash = len(path)
			}
			segment := path[:slash]
			path = path[slash:]
			switch p := vars[0].(type) {
			case *string:
				*p = segment
			case *int64:
				n, err := strconv.ParseInt(segment, 10, 64)
				if err != nil || n < 0 {
					return false
				}
				*p = n
			case *int:
				n, err := strconv.Atoi(segment)
				if err != nil || n < 0 {
					return false
				}
				*p = n
			default:
				panic("vars must be *string or *int")
			}
			vars = vars[1:]
		case path[0]:
			// non-'+' pattern byte must match path byte
			path = path[1:]
		default:
			return false
		}
	}
	return path == "" && pattern == ""
}


// Process hash generation service
func HashHandler(w http.ResponseWriter, r *http.Request)  {
    if !shutdown {
        reqNum := IncRequests()
        currentAvg := GetAverage()
        start := time.Now()
        w.Header().Set("Content-Type", "text/plain")
        switch r.Method {
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
            map_of_input[reqNum] = model.NewTimedData(reqNum, pwParam)
            fmt.Fprintf(w, "%d\n", reqNum)

        default:
            fmt.Fprintf(w, "Sorry, only POST methods are supported.")
        }
        executeTime := time.Now().Sub(start).Microseconds()
        IncAverage( reqNum, currentAvg, executeTime)
    }
}

