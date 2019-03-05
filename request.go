package main

import "encoding/json"

type RequestParams struct {
    Command string          `json:"command"`
    Value   json.RawMessage `json:"value"`
}
