package main

import (

    //"fmt"

    "encoding/json"
    "io"
    "log"
)

const (
    succeeded = true
    failed    = false
)

type Response struct {
    Command string      `json:"command"`
    Success bool        `json:"success"`
    Value   interface{} `json:"value"`
}

func (r *Response) write(w io.Writer) {
    encoder := json.NewEncoder(w)
    err := encoder.Encode(r)
    if err != nil {
        log.Panic(err)
    }
}
