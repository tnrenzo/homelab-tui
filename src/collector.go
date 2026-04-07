package main

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

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

	// hardware
	CPU      int    `json:"cpu"`
	CPUL     int    `json:"cpu_logical"`
	MemTotal uint64 `json:"totalmem"`
	MemFree  uint64 `json:"freemem"`

	Disks []DiskInfo `json:"disks"`
}

func (s *SystemInfo) fetchProc() error {
	h, err := host.Info()
	if err != nil {
		return err
	}

	m, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	uptime, err := host.Uptime()
	if err != nil {
		return err
	}

	physicalc, err := cpu.Counts(false)
	if err != nil {
		return err
	}

	logicalc, err := cpu.Counts(true)
	if err != nil {
		return err
	}

	partitions, err := disk.Partitions(false)
	if err != nil {
		return err
	}

	seen := map[string]bool{}
	for _, p := range partitions {
		if seen[p.Device] {
			continue // skip duplicates
		}
		seen[p.Device] = true

		u, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		s.Disks = append(s.Disks, DiskInfo{
			Mountpoint: p.Mountpoint,
			Total:      u.Total,
			Used:       u.Used,
			Free:       u.Free,
			UsedPct:    u.UsedPercent,
		})
	}

	s.Hostname = h.Hostname
	s.Arch = h.KernelArch
	s.KVersion = h.KernelVersion
	s.Uptime = uptime
	s.CPU = physicalc
	s.CPUL = logicalc
	s.MemTotal = m.Total
	s.MemFree = m.Free

	return nil
}
