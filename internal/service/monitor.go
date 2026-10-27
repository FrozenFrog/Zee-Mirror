package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"zee-mirror/internal/toolversion"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type SystemStats struct {
	StartTime   time.Time
	Uptime      time.Duration
	CPUUsage    float64
	DiskUsage   float64
	MemoryTotal uint64
	MemoryUsed  uint64
	MemoryFree  uint64
	MemoryBot   uint64
	DiskTotal   uint64
	DiskUsed    uint64
	DiskFree    uint64
	GoRoutines  int
	ActiveTasks int
	QueuedTasks int
}

var botStartTime = time.Now()

func (s *BotService) GetSystemStats() SystemStats {
	stats := SystemStats{
		StartTime:   botStartTime,
		Uptime:      time.Since(botStartTime),
		GoRoutines:  runtime.NumGoroutine(),
		ActiveTasks: len(s.TaskManager.GetActiveTasks()),
	}

	if v, err := mem.VirtualMemory(); err == nil {
		stats.MemoryTotal = v.Total
		stats.MemoryUsed = v.Used
		stats.MemoryFree = v.Available
	}
	if p, err := process.NewProcess(int32(os.Getpid())); err == nil { //nolint:gosec // G115: PIDs fit in int32
		if info, err := p.MemoryInfo(); err == nil {
			stats.MemoryBot = info.RSS
		}
	}

	diskTotal, diskUsed, diskFree := s.GetDiskStats(s.Config.DownloadDir)
	stats.DiskTotal = diskTotal
	stats.DiskUsed = diskUsed
	stats.DiskFree = diskFree
	if diskTotal > 0 {
		stats.DiskUsage = float64(diskUsed) / float64(diskTotal) * 100
	}

	stats.QueuedTasks = s.TaskManager.Queue.Count()

	return stats
}

// GetDiskStats reports the filesystem holding path, falling back to the root
// filesystem when path does not exist yet.
func (s *BotService) GetDiskStats(path string) (total, used, free uint64) {
	u, err := disk.Usage(path)
	if err != nil {
		if u, err = disk.Usage("/"); err != nil {
			return 0, 0, 0
		}
	}
	return u.Total, u.Used, u.Free
}

func (s *BotService) getDiskUsageOS() float64 {
	total, used, _ := s.GetDiskStats(s.Config.DownloadDir)
	if total == 0 {
		return 0
	}
	return float64(used) / float64(total) * 100
}

func percent(part, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

// FormatSystemStats is FormatSystemStatsLang in DefaultLang.
func (s *BotService) FormatSystemStats(stats SystemStats) string {
	return s.FormatSystemStatsLang(i18n.DefaultLang, stats)
}

// FormatSystemStatsLang renders stats as a MarkdownV2 message in lang.
func (s *BotService) FormatSystemStatsLang(lang string, stats SystemStats) string {
	var content strings.Builder
	line := func(key string, args ...any) {
		content.WriteString(i18n.T(lang, key, args...))
		content.WriteString("\n")
	}

	line("monitor_uptime_header")
	line("monitor_uptime", utils.EscapeMarkdownV2(formatUptime(stats.Uptime)))
	line("monitor_started_at", utils.EscapeMarkdownV2(stats.StartTime.Format("02 Jan 15:04")))
	fmt.Fprintf(&content, "• Goroutines: `%d`\n\n", stats.GoRoutines)

	line("monitor_tasks_header")
	line("monitor_tasks_running", stats.ActiveTasks)
	line("monitor_tasks_queued", stats.QueuedTasks)
	content.WriteString("\n")

	line("monitor_memory_header")
	memBar := createUsageBar(percent(stats.MemoryUsed, stats.MemoryTotal), 12)
	line("monitor_used",
		utils.EscapeMarkdownV2(utils.FormatBytesUint64(stats.MemoryUsed)),
		utils.EscapeMarkdownV2(utils.FormatBytesUint64(stats.MemoryTotal)))
	line("monitor_bot_memory", utils.EscapeMarkdownV2(utils.FormatBytesUint64(stats.MemoryBot)))
	fmt.Fprintf(&content, "%s\n\n", utils.EscapeMarkdownV2(memBar))

	line("monitor_disk_header")
	diskBar := createUsageBar(stats.DiskUsage, 12)
	line("monitor_used",
		utils.EscapeMarkdownV2(utils.FormatBytesUint64(stats.DiskUsed)),
		utils.EscapeMarkdownV2(utils.FormatBytesUint64(stats.DiskTotal)))
	content.WriteString(utils.EscapeMarkdownV2(diskBar))

	return ProfessionalMessage(i18n.T(lang, "monitor_stats_title"), content.String())
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm %ds", minutes, int(d.Seconds())%60)
}

func createUsageBar(percentage float64, width int) string {
	filled := int(percentage * float64(width) / 100)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	icon := "🟢"
	if percentage > 70 {
		icon = "🟡"
	}
	if percentage > 90 {
		icon = "🔴"
	}

	return fmt.Sprintf("%s %s %.1f%%", icon, bar, percentage)
}

type HealthCheck struct {
	Name           string
	Status         string
	Message        string
	Healthy        bool
	CurrentVersion string
	LatestVersion  string
}

// PerformHealthChecks is PerformHealthChecksLang in DefaultLang.
func (s *BotService) PerformHealthChecks() []HealthCheck {
	return s.PerformHealthChecksLang(i18n.DefaultLang)
}

// PerformHealthChecksLang runs the health checks, with names and messages
// (plain text) in lang.
func (s *BotService) PerformHealthChecksLang(lang string) []HealthCheck {
	var checks []HealthCheck
	versionCtx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	versions := toolversion.All(versionCtx)
	cancel()

	aria2Check := HealthCheck{Name: "Aria2", CurrentVersion: versions["aria2"].Current, LatestVersion: versions["aria2"].Latest}
	if _, err := os.Stat("/usr/bin/aria2c"); err == nil {
		aria2Check.Status = IconOK
		aria2Check.Message = i18n.T(lang, "monitor_aria2_found")
		aria2Check.Healthy = true
	} else {
		aria2Check.Status = "⚠️ WARNING"
		aria2Check.Message = i18n.T(lang, "monitor_aria2_missing")
		aria2Check.Healthy = true
	}
	checks = append(checks, aria2Check)

	rcloneCheck := HealthCheck{Name: "Rclone", CurrentVersion: versions["rclone"].Current, LatestVersion: versions["rclone"].Latest}
	if _, err := os.Stat(s.TaskManager.ConfigDir + "/rclone.conf"); err == nil {
		rcloneCheck.Status = IconOK
		rcloneCheck.Message = i18n.T(lang, "monitor_rclone_ok")
		rcloneCheck.Healthy = true
	} else {
		rcloneCheck.Status = IconError
		rcloneCheck.Message = i18n.T(lang, "monitor_rclone_missing")
		rcloneCheck.Healthy = false
	}
	checks = append(checks, rcloneCheck)
	ytdlpCheck := HealthCheck{Name: "yt-dlp", CurrentVersion: versions["yt-dlp"].Current, LatestVersion: versions["yt-dlp"].Latest}
	if ytdlpCheck.CurrentVersion != "" {
		ytdlpCheck.Status = IconOK
		ytdlpCheck.Message = i18n.T(lang, "monitor_ytdlp_found")
		ytdlpCheck.Healthy = true
	} else {
		ytdlpCheck.Status = IconError
		ytdlpCheck.Message = i18n.T(lang, "monitor_ytdlp_missing")
	}
	checks = append(checks, ytdlpCheck)

	diskCheck := HealthCheck{Name: i18n.T(lang, "monitor_disk_name")}
	diskUsage := s.getDiskUsageOS()
	switch {
	case diskUsage < 85:
		diskCheck.Status = IconOK
		diskCheck.Message = i18n.T(lang, "monitor_disk_ok", diskUsage)
		diskCheck.Healthy = true
	case diskUsage < 95:
		diskCheck.Status = "⚠️ WARNING"
		diskCheck.Message = i18n.T(lang, "monitor_disk_warn", diskUsage)
		diskCheck.Healthy = true
	default:
		diskCheck.Status = "❌ CRITICAL"
		diskCheck.Message = i18n.T(lang, "monitor_disk_critical", diskUsage)
		diskCheck.Healthy = false
	}
	checks = append(checks, diskCheck)

	dbCheck := HealthCheck{Name: "Database"}
	ctx := context.Background()
	if err := s.DB.Ping(ctx); err == nil {
		dbCheck.Status = IconOK
		dbCheck.Message = i18n.T(lang, "monitor_db_ok")
		dbCheck.Healthy = true
	} else {
		dbCheck.Status = IconError
		dbCheck.Message = i18n.T(lang, "monitor_db_error")
		dbCheck.Healthy = false
	}
	checks = append(checks, dbCheck)

	dlCheck := HealthCheck{Name: i18n.T(lang, "monitor_dl_name")}
	if info, err := os.Stat(s.Config.DownloadDir); err == nil && info.IsDir() {
		dlCheck.Status = IconOK
		dlCheck.Message = i18n.T(lang, "monitor_dl_ok")
		dlCheck.Healthy = true
	} else {
		dlCheck.Status = IconError
		dlCheck.Message = i18n.T(lang, "monitor_dl_error")
		dlCheck.Healthy = false
	}
	checks = append(checks, dlCheck)

	return checks
}

// FormatHealthCheck is FormatHealthCheckLang in DefaultLang.
func (s *BotService) FormatHealthCheck(checks []HealthCheck) string {
	return s.FormatHealthCheckLang(i18n.DefaultLang, checks)
}

// FormatHealthCheckLang renders checks as a MarkdownV2 message in lang.
func (s *BotService) FormatHealthCheckLang(lang string, checks []HealthCheck) string {
	var content strings.Builder

	allHealthy := true
	for _, c := range checks {
		if !c.Healthy {
			allHealthy = false
			break
		}
	}

	status := i18n.T(lang, "monitor_health_ok")
	if !allHealthy {
		status = i18n.T(lang, "monitor_health_issue")
	}

	fmt.Fprintf(&content, "%s\n%s\n\n", i18n.T(lang, "monitor_health_status", status), CompactSeparator)

	for _, c := range checks {
		icon := "✅"
		if !c.Healthy {
			if strings.Contains(c.Status, "WARNING") {
				icon = "⚠️"
			} else {
				icon = "❌"
			}
		}

		fmt.Fprintf(&content, "%s *%s*\n", icon, utils.EscapeMarkdownV2(c.Name))
		fmt.Fprintf(&content, "└ _%s_\n", utils.EscapeMarkdownV2(c.Message))
		if c.Name == "Aria2" || c.Name == "Rclone" || c.Name == "yt-dlp" {
			current, latest := c.CurrentVersion, c.LatestVersion
			if current == "" {
				current = "—"
			}
			if latest == "" {
				latest = "—"
			}
			fmt.Fprintf(&content, "└ %s\n", i18n.T(lang, "monitor_version_current", utils.EscapeMarkdownV2(current)))
			fmt.Fprintf(&content, "└ %s\n", i18n.T(lang, "monitor_version_latest", utils.EscapeMarkdownV2(latest)))
		}
	}

	fmt.Fprintf(&content, "\n%s\n%s", CompactSeparator, i18n.T(lang, "monitor_checked_at", utils.EscapeMarkdownV2(time.Now().Format("15:04:05"))))

	return ProfessionalMessage(i18n.T(lang, "monitor_health_title"), content.String())
}

func (s *BotService) PerformDiskCleanup() {
	slog.Info("Starting disk cleanup")

	files, err := os.ReadDir(os.TempDir())
	if err == nil {
		for _, f := range files {
			if strings.HasPrefix(f.Name(), "zee-mirror") {
				_ = os.Remove(filepath.Join(os.TempDir(), f.Name()))
			}
		}
	}
}

func (s *BotService) StartResourceMonitor() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-s.TaskManager.ShutdownChan:
				return
			case <-ticker.C:
				s.checkResourceAlerts()
			}
		}
	}()

	slog.Info("Resource monitoring started")
}

func (s *BotService) checkResourceAlerts() {
	diskUsage := s.getDiskUsageOS()

	if diskUsage > 90 && s.Notifications != nil {
		s.Notifications.AlertDiskUsage(diskUsage, s.Config.DownloadDir)
	}

	v, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	if memUsage := v.UsedPercent; memUsage > 90 && s.Notifications != nil {
		s.Notifications.AlertSystemHealth(0, memUsage)
	}
}
