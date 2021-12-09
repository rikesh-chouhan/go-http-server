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
)


func main() {

    start := time.Now()
    end := time.Now()

}

func SendPostAsync(url string, body []byte, rc chan *http.Response) error {
    response, err := http.Post(url, "application/json", bytes.NewReader(body))
    if err == nil {
        rc <- response
    }

    return err
}
