package main

import (
	"tn-renzo/homelab-tui/shared/protocol"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

func fetchSystemInfo() (protocol.SystemInfo, error) {
	s := protocol.SystemInfo{}

	h, err := host.Info()
	if err != nil {
		return s, err
	}

	m, err := mem.VirtualMemory()
	if err != nil {
		return s, err
	}

	uptime, err := host.Uptime()
	if err != nil {
		return s, err
	}

	physicalc, err := cpu.Counts(false)
	if err != nil {
		return s, err
	}

	logicalc, err := cpu.Counts(true)
	if err != nil {
		return s, err
	}

	core_usage, err := cpu.Percent(0, true)
	if err != nil {
		return s, err
	}

	partitions, err := disk.Partitions(false)
	if err != nil {
		return s, err
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
		s.Disks = append(s.Disks, protocol.DiskInfo{
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
	s.CoreUsage = core_usage
	s.MemTotal = m.Total
	s.MemUsed = m.Used
	s.MemFree = m.Free

	return s, nil
}
