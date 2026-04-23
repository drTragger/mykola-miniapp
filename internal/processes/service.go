package processes

import (
	"runtime"
	"sort"
	"strings"
	"sync"
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
	infos := buildProcessInfosParallel(procs, now)

	items := make([]ProcessInfo, 0, len(infos))
	search := strings.ToLower(strings.TrimSpace(params.Search))

	for _, info := range infos {
		if info.PID == 0 {
			continue
		}
		if search != "" && !matchesSearch(info, search) {
			continue
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

func buildProcessInfosParallel(procs []*process.Process, now time.Time) []ProcessInfo {
	infos := make([]ProcessInfo, len(procs))

	workers := runtime.GOMAXPROCS(0) * 2
	if workers < 4 {
		workers = 4
	}
	if workers > len(procs) {
		workers = len(procs)
	}
	if workers < 1 {
		return infos
	}

	jobs := make(chan int, len(procs))
	for i := range procs {
		jobs <- i
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for idx := range jobs {
				info, ok := buildProcessInfo(procs[idx], now)
				if ok {
					infos[idx] = info
				}
			}
		}()
	}

	wg.Wait()
	return infos
}

func matchesSearch(info ProcessInfo, q string) bool {
	return strings.Contains(strings.ToLower(info.Name), q) ||
		strings.Contains(strings.ToLower(info.Cmdline), q) ||
		strings.Contains(strings.ToLower(info.User), q)
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

	var less func(i, j int) bool

	switch strings.ToLower(sortBy) {
	case "memory":
		less = func(i, j int) bool { return items[i].RSSBytes > items[j].RSSBytes }
	case "name":
		less = func(i, j int) bool { return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name) }
	case "pid":
		less = func(i, j int) bool { return items[i].PID < items[j].PID }
	case "uptime":
		less = func(i, j int) bool { return items[i].UptimeSec > items[j].UptimeSec }
	default:
		less = func(i, j int) bool { return items[i].CPUPercent > items[j].CPUPercent }
	}

	sort.Slice(items, func(i, j int) bool {
		if desc {
			return less(i, j)
		}
		return !less(i, j)
	})
}

var criticalCmdlineParts = []string{
	"systemd",
	"sshd",
	"dbus",
	"cloudflared",
	"wg-quick",
	"mykola-miniapp",
	"mykola-bot",
}

func isCriticalProcess(pid int32, name, cmdline string) bool {
	if pid <= 2 {
		return true
	}

	if _, ok := criticalNames[name]; ok {
		return true
	}

	lc := strings.ToLower(cmdline)
	for _, part := range criticalCmdlineParts {
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
