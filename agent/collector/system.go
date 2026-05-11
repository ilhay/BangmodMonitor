package collector

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type Metrics struct {
	Timestamp int64      `json:"timestamp"`
	Hostname  string     `json:"hostname"`
	CPU       CPUMetric  `json:"cpu"`
	Memory    MemMetric  `json:"memory"`
	Disks     []DiskMetric    `json:"disks"`
	Networks  []NetMetric     `json:"networks"`
	Apps      AppStatus  `json:"apps"`
	Processes []ProcMetric `json:"processes"`
}

type CPUMetric struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
}

type MemMetric struct {
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskMetric struct {
	Path         string  `json:"path"`
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

type NetMetric struct {
	Interface   string `json:"interface"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

type AppStatus struct {
	Nginx      bool `json:"nginx"`
	Apache     bool `json:"apache"`
	MySQL      bool `json:"mysql"`
	MariaDB    bool `json:"mariadb"`
	PostgreSQL bool `json:"postgresql"`
}

type ProcMetric struct {
	Name          string  `json:"name"`
	PID           int32   `json:"pid"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	Status        string  `json:"status"`
}

func Collect(hostname string) (*Metrics, error) {
	m := &Metrics{
		Timestamp: time.Now().Unix(),
		Hostname:  hostname,
	}

	if percents, err := cpu.Percent(time.Second, false); err == nil && len(percents) > 0 {
		counts, _ := cpu.Counts(true)
		m.CPU = CPUMetric{UsagePercent: percents[0], Cores: counts}
	}

	if v, err := mem.VirtualMemory(); err == nil {
		m.Memory = MemMetric{
			TotalBytes:   v.Total,
			UsedBytes:    v.Used,
			UsagePercent: v.UsedPercent,
		}
	}

	if parts, err := disk.Partitions(false); err == nil {
		for _, p := range parts {
			if usage, err := disk.Usage(p.Mountpoint); err == nil {
				m.Disks = append(m.Disks, DiskMetric{
					Path:         p.Mountpoint,
					TotalBytes:   usage.Total,
					UsedBytes:    usage.Used,
					UsagePercent: usage.UsedPercent,
				})
			}
		}
	}

	if ifaces, err := net.IOCounters(true); err == nil {
		for _, iface := range ifaces {
			m.Networks = append(m.Networks, NetMetric{
				Interface:   iface.Name,
				BytesSent:   iface.BytesSent,
				BytesRecv:   iface.BytesRecv,
				PacketsSent: iface.PacketsSent,
				PacketsRecv: iface.PacketsRecv,
			})
		}
	}

	m.Apps = collectApps()
	m.Processes = collectProcesses()

	return m, nil
}

func collectApps() AppStatus {
	procs, _ := process.Processes()
	status := AppStatus{}
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		switch name {
		case "nginx":
			status.Nginx = true
		case "apache2", "httpd":
			status.Apache = true
		case "mysqld":
			status.MySQL = true
		case "mariadbd", "mysqld_safe":
			status.MariaDB = true
		case "postgres":
			status.PostgreSQL = true
		}
	}
	return status
}

func collectProcesses() []ProcMetric {
	procs, err := process.Processes()
	if err != nil {
		return nil
	}
	var result []ProcMetric
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		cpu, _ := p.CPUPercent()
		mem, _ := p.MemoryPercent()
		statuses, _ := p.Status()
		status := "unknown"
		if len(statuses) > 0 {
			status = statuses[0]
		}
		result = append(result, ProcMetric{
			Name:          name,
			PID:           p.Pid,
			CPUPercent:    cpu,
			MemoryPercent: mem,
			Status:        status,
		})
		if len(result) >= 50 {
			break
		}
	}
	return result
}
