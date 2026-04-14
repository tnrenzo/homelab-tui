package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coder/websocket"
)

type Payload struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// creates the websocket and sends system data
func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		fmt.Fprint(os.Stdout, err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx := context.Background()

	// add a rate limit
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		info := SystemInfo{}
		if err := info.fetchProc(); err != nil {
			fmt.Fprint(os.Stderr, "Error fetching Info", err)
			return
		}

		data, err := json.Marshal(info)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		msg := Payload{
			Type:    "system_info",
			Payload: json.RawMessage(data),
		}

		out, err := json.Marshal(msg)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		err = conn.Write(ctx, websocket.MessageText, out)
		if err != nil {
			return
		}
	}
}
