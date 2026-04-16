package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

/* structs copied across backend and frontend for now - fix later */
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SystemInfo struct {
	// host
	Hostname string `json:"hostname"`
	Arch     string `json:"arch"`
	KVersion string `json:"kernel"`
	Uptime   uint64 `json:"uptime"`

	// hardware
	CPU      int    `json:"cpu"`
	CPUL     int    `json:"cpu_logical"`
	MemTotal uint64 `json:"totalmem"`
	MemFree  uint64 `json:"freemem"`

	Disks []DiskInfo `json:"disks"`
}

type DiskInfo struct {
	Mountpoint string  `json:"mountpoint"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Free       uint64  `json:"free"`
	UsedPct    float64 `json:"used_pct"`
}

// latestSystemInfo is updated as websocket messages arrive.
var latestSystemInfo SystemInfo
var latestSystemInfoMu sync.RWMutex

func setLatestSystemInfo(info SystemInfo) {
	latestSystemInfoMu.Lock()
	latestSystemInfo = info
	latestSystemInfoMu.Unlock()
}

func getLatestSystemInfo() SystemInfo {
	latestSystemInfoMu.RLock()
	defer latestSystemInfoMu.RUnlock()
	return latestSystemInfo
}

// connect to ws and receive data
func handleWS() {
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, "ws://localhost:8080/ws", nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	for {
		var msg Message
		err := wsjson.Read(ctx, conn, &msg)
		if err != nil {
			fmt.Println("read error:", err)
			return
		}

		var info SystemInfo
		if err := json.Unmarshal(msg.Payload, &info); err != nil {
			fmt.Println("decode error:", err)
			return
		}

		setLatestSystemInfo(info)
	}
}
