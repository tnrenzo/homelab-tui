package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	fmt.Println("WS on 0.0.0.0:10001")
	http.HandleFunc("/ws", handleWS)
	err := http.ListenAndServe(":10001", nil)
	if err != nil {
		fmt.Fprint(os.Stderr, "Error creating websocket", err)
	}
}
