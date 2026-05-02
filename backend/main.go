package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func main() {
	addr := flag.String("addr", "127.0.0.1", "<host>")
	port := flag.String("port", "8080", "<port>")

	flag.Parse()

	wsPort, err := strconv.Atoi(*port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid port: %v\n", err)
		return
	}

	if wsPort > 65535 || wsPort < 1 {
		fmt.Fprintf(os.Stderr, "Invalid Port: %d\n", wsPort)
		return
	}

	listenAddr := *addr + ":" + strconv.Itoa(wsPort)
	fmt.Printf("WS on %s\n", listenAddr)

	http.HandleFunc("/ws", handleWS)
	err = http.ListenAndServe(listenAddr, nil)
	if err != nil {
		fmt.Fprint(os.Stderr, "Error creating websocket: ", err)
	}
}
