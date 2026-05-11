package protocol

import (
	"encoding/json"
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type DiskInfo struct {
	Mountpoint string  `json:"mountpoint"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Free       uint64  `json:"free"`
	UsedPct    float64 `json:"used_pct"`
}

type ProcessInfo struct {
	Name string `json:"process"`
	User string `json:"user"`
	Pid  int    `json:"pid"`
}

type NetworkInfo struct {
	Interface string `json:"interface"`
	RxBytes   uint64 `json:"rx_bytes"`
	TxBytes   uint64 `json:"tx_bytes"`
}

type LoadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

type SystemInfo struct {
	// host
	Hostname string `json:"hostname"`
	Arch     string `json:"arch"`
	KVersion string `json:"kernel"`
	Uptime   uint64 `json:"uptime"`

	// CPU
	CPU       int       `json:"cpu"`
	CPUL      int       `json:"cpu_logical"`
	CoreUsage []float64 `json:"core_usage"`

	Load []LoadInfo `json:"load"`

	// Memory
	MemTotal uint64 `json:"totalmem"`
	MemUsed  uint64 `json:"usedmem"`
	MemFree  uint64 `json:"freemem"`

	Disks []DiskInfo `json:"disks"`

	Processes []ProcessInfo `json:"procs"`

	Network []NetworkInfo `json:"network"`
}
