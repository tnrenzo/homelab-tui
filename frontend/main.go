package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	tea "charm.land/bubbletea/v2"
)

func main() {
	addr := flag.String("addr", "127.0.0.1", "<addr>")
	port := flag.String("port", "8080", "<port>")

	flag.Parse()

	wsPort, err := strconv.Atoi(*port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid port: %v\n", err)
		return
	}

	if wsPort > 65535 || wsPort < 1 {
		fmt.Fprintf(os.Stdout, "Invalid Port: %d\n", wsPort)
		return
	}

	wsAddr := *addr

	p := tea.NewProgram(newModel(wsAddr, wsPort))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
