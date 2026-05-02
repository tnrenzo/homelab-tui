package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"tn-renzo/homelab-tui/shared/protocol"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Message = protocol.Message
type SystemInfo = protocol.SystemInfo
type DiskInfo = protocol.DiskInfo

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
func handleWS(addr string, port int) {
	WSURL := "ws://" + addr + ":" + strconv.Itoa(port) + "/ws"
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, WSURL, nil)
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
