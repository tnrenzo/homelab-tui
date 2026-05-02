package protocol

import "encoding/json"

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

	// Memory
	MemTotal uint64 `json:"totalmem"`
	MemUsed  uint64 `json:"usedmem"`
	MemFree  uint64 `json:"freemem"`

	Disks []DiskInfo `json:"disks"`
}
