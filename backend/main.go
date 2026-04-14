package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	fmt.Println("WS on 127.0.0.1:8080")
	http.HandleFunc("/ws", handleWS)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Fprint(os.Stderr, "Error creating websocket", err)
	}
}
