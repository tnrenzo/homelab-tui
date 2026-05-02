package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func main() {
	/* TODO: Add CLI parsing */
	wsAddr := "127.0.0.1"
	wsPort := "8080"

	addrInput := ""
	portInput := ""

	fmt.Printf("WebSocket IP [%s]: ", wsAddr)
	if _, err := fmt.Scanln(&addrInput); err == nil {
		addrInput = strings.TrimSpace(addrInput)
		if addrInput != "" {
			wsAddr = addrInput
		}
	}

	fmt.Printf("WebSocket port [%s]: ", wsPort)
	if _, err := fmt.Scanln(&portInput); err == nil {
		portInput = strings.TrimSpace(portInput)
		if portInput != "" {
			wsPort = portInput
		}
	}

	p := tea.NewProgram(newModel(wsAddr, wsPort))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
