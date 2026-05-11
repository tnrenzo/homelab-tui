package main

import (
	"tn-renzo/homelab-tui/shared/protocol"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
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

	procs, err := process.Processes()
	if err != nil {
		return s, err
	}
	for _, p := range procs {
		name, _ := p.Name()
		username, _ := p.Username()
		pid := p.Pid
		s.Processes = append(s.Processes, protocol.ProcessInfo{
			Name: name,
			User: username,
			Pid:  int(pid),
		})
	}

	lo, err := load.Avg()
	if err != nil {
		return s, err
	}
	s.Load = append(s.Load, protocol.LoadInfo{
		Load1:  lo.Load1,
		Load5:  lo.Load5,
		Load15: lo.Load15,
	})

	iocounters, err := net.IOCounters(true)
	if err != nil {
		return s, err
	}
	for _, io := range iocounters {
		s.Network = append(s.Network, protocol.NetworkInfo{
			Interface: io.Name,
			RxBytes:   io.BytesRecv,
			TxBytes:   io.BytesSent,
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
