package main

import (
    "time"
    "crypto/sha512"
)

typedef Hashed struct {
    Start Time
    Count int64
    Text, Hashed_Text string
}

typedef Stats struct {
    Total, Average int64
}

func MakeHashed(count int64, text string) Hashed {
    return Hashed {
        Start: time.Now(),
        Count: count,
        Text: text,
        Hashed_Text: GenerateSha512(text),
    }
}

func GenerateSha512(text string) string {
    return sha512.Sum512([]byte (text))
}
