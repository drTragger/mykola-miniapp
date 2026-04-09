package processes

import (
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

type ListParams struct {
	Search string
	Sort   string
	Dir    string
	Limit  int
	Offset int
}

var criticalNames = map[string]struct{}{
	"systemd":          {},
	"sshd":             {},
	"dbus-daemon":      {},
	"systemd-journald": {},
	"systemd-logind":   {},
	"systemd-udevd":    {},
	"cron":             {},
	"rsyslogd":         {},
	"mykola-miniapp":   {},
	"mykola-bot":       {},
	"cloudflared":      {},
	"wg-quick":         {},
}

func ListProcesses(params ListParams) ([]ProcessInfo, int, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, 0, err
	}

	now := time.Now()
	items := make([]ProcessInfo, 0, len(procs))

	for _, p := range procs {
		info, ok := buildProcessInfo(p, now)
		if !ok {
			continue
		}

		if params.Search != "" {
			q := strings.ToLower(params.Search)
			if !strings.Contains(strings.ToLower(info.Name), q) &&
				!strings.Contains(strings.ToLower(info.Cmdline), q) &&
				!strings.Contains(strings.ToLower(info.User), q) {
				continue
			}
		}

		items = append(items, info)
	}

	total := len(items)
	sortProcesses(items, params.Sort, params.Dir)

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= len(items) {
		return []ProcessInfo{}, total, nil
	}

	end := offset + limit
	if end > len(items) {
		end = len(items)
	}

	return items[offset:end], total, nil
}

func buildProcessInfo(p *process.Process, now time.Time) (ProcessInfo, bool) {
	pid := p.Pid

	name, _ := p.Name()
	if name == "" {
		name = "unknown"
	}

	ppid, _ := p.Ppid()
	user, _ := p.Username()
	statusSlice, _ := p.Status()
	status := ""
	if len(statusSlice) > 0 {
		status = statusSlice[0]
	}

	cpuPercent, _ := p.CPUPercent()
	memPercent, _ := p.MemoryPercent()
	memInfo, _ := p.MemoryInfo()
	threads, _ := p.NumThreads()
	cmdline, _ := p.Cmdline()
	exe, _ := p.Exe()
	createTimeMs, _ := p.CreateTime()

	var rss, vms uint64
	if memInfo != nil {
		rss = memInfo.RSS
		vms = memInfo.VMS
	}

	createTime := time.UnixMilli(createTimeMs)
	uptimeSec := int64(0)
	if !createTime.IsZero() && createTime.Before(now) {
		uptimeSec = int64(now.Sub(createTime).Seconds())
	}

	return ProcessInfo{
		PID:           pid,
		PPID:          ppid,
		Name:          name,
		User:          user,
		Status:        status,
		CPUPercent:    cpuPercent,
		MemoryPercent: memPercent,
		RSSBytes:      rss,
		VMSBytes:      vms,
		Threads:       threads,
		Cmdline:       cmdline,
		Executable:    exe,
		CreateTime:    createTimeMs,
		UptimeSec:     uptimeSec,
		IsCritical:    isCriticalProcess(pid, name, cmdline),
		ServiceName:   detectServiceName(name, cmdline),
	}, true
}

func sortProcesses(items []ProcessInfo, sortBy, dir string) {
	desc := strings.ToLower(dir) != "asc"

	less := func(i, j int) bool { return items[i].CPUPercent > items[j].CPUPercent }

	switch strings.ToLower(sortBy) {
	case "memory":
		less = func(i, j int) bool { return items[i].RSSBytes > items[j].RSSBytes }
	case "name":
		less = func(i, j int) bool { return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name) }
	case "pid":
		less = func(i, j int) bool { return items[i].PID < items[j].PID }
	case "uptime":
		less = func(i, j int) bool { return items[i].UptimeSec > items[j].UptimeSec }
	}

	sort.Slice(items, func(i, j int) bool {
		if desc {
			return less(i, j)
		}
		return !less(i, j)
	})
}

func isCriticalProcess(pid int32, name, cmdline string) bool {
	if pid <= 2 {
		return true
	}

	if _, ok := criticalNames[name]; ok {
		return true
	}

	lc := strings.ToLower(cmdline)

	for _, part := range []string{
		"systemd",
		"sshd",
		"dbus",
		"cloudflared",
		"wg-quick",
		"mykola-miniapp",
		"mykola-bot",
	} {
		if strings.Contains(lc, part) {
			return true
		}
	}

	return false
}

func detectServiceName(name, cmdline string) string {
	lc := strings.ToLower(cmdline)

	switch {
	case strings.Contains(lc, "jellyfin"):
		return "jellyfin"
	case strings.Contains(lc, "qbittorrent"):
		return "qbittorrent"
	case strings.Contains(lc, "sonarr"):
		return "sonarr"
	case strings.Contains(lc, "radarr"):
		return "radarr"
	case strings.Contains(lc, "prowlarr"):
		return "prowlarr"
	case strings.Contains(lc, "cloudflared"):
		return "cloudflared"
	default:
		return name
	}
}

func SignalProcess(pid int32, action string) error {
	if pid <= 2 {
		return syscall.EPERM
	}

	exists, err := process.PidExists(pid)
	if err != nil || !exists {
		return syscall.ESRCH
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}

	name, _ := p.Name()
	cmdline, _ := p.Cmdline()

	if isCriticalProcess(pid, name, cmdline) {
		return syscall.EPERM
	}

	var sig syscall.Signal

	switch strings.ToLower(action) {
	case "term":
		sig = syscall.SIGTERM
	case "kill":
		sig = syscall.SIGKILL
	case "stop":
		sig = syscall.SIGSTOP
	case "cont":
		sig = syscall.SIGCONT
	default:
		return syscall.EINVAL
	}

	return syscall.Kill(int(pid), sig)
}
