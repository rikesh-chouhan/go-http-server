package main

import (
    "time"
    "crypto/sha512"
    "fmt"
)

type TimedData struct {
    Start time.Time
    Count int64
    Text string
}

type Stats struct {
    Total, Average int64
}

func NewTimedData(count int64, text string) TimedData {
    return Hashed {
        Start: time.Now(),
        Count: count,
        Text: text,
    }
}

func GenerateHash(timedData *TimedData) string {
    return fmt.Sprintf("%x", sha512.Sum512([]byte (timedData.text)))
}
