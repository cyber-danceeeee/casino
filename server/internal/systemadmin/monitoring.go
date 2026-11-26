package systemadmin

import (
	"context"
	"runtime"
	"syscall"
	"time"

	"casino/internal/database"
)

var applicationStartedAt = time.Now()

func ApplicationStartedAtForHealth() time.Time { return applicationStartedAt }

func (s *Service) Monitoring() (map[string]any, error) {
	current := now()
	services := []map[string]any{}

	databaseStatus := map[string]any{"service": "database", "status": "DOWN"}
	if sqlDB, err := s.db.DB(); err == nil {
		start := time.Now()
		err = sqlDB.Ping()
		stats := sqlDB.Stats()
		databaseStatus = map[string]any{"service": "database", "status": statusOf(err), "latencyMS": time.Since(start).Milliseconds(), "openConnections": stats.OpenConnections, "inUseConnections": stats.InUse, "idleConnections": stats.Idle}
	}
	services = append(services, databaseStatus)

	redisStatus := map[string]any{"service": "redis", "status": "DOWN"}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	start := time.Now()
	err := database.GetRedisCli().Ping(ctx).Err()
	cancel()
	redisStatus = map[string]any{"service": "redis", "status": statusOf(err), "latencyMS": time.Since(start).Milliseconds()}
	services = append(services, redisStatus)

	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	disk := diskUsage(".")
	var activeSessions, failedTasks, pendingJobs int64
	_ = s.db.Table("game_admin_sessions").Where("revoked_at IS NULL AND expired_at>? AND deleted_at IS NULL", current).Count(&activeSessions).Error
	_ = s.db.Table("game_system_task_executions").Where("status='FAILED' AND started_at>=? AND deleted_at IS NULL", current.Add(-24*time.Hour)).Count(&failedTasks).Error
	_ = s.db.Table("game_notification_jobs").Where("status IN ('PENDING','PROCESSING') AND deleted_at IS NULL").Count(&pendingJobs).Error

	providers := []map[string]any{}
	_ = s.db.Table("game_game_providers").Select("provider_code provider,name,status enabled,api_status status,last_tested_at lastTestedAt,last_sync_at lastSyncAt").Where("deleted_at IS NULL").Order("provider_code").Find(&providers).Error

	callback := struct{ Total, Success int64 }{}
	_ = s.db.Table("game_deposit_callback_logs").Select("COUNT(*) total,SUM(CASE WHEN processing_result='SUCCESS' THEN 1 ELSE 0 END) success").Where("created_at>=? AND deleted_at IS NULL", current.Add(-time.Hour)).Scan(&callback).Error
	callbackRate := any(nil)
	if callback.Total > 0 {
		callbackRate = float64(callback.Success) * 100 / float64(callback.Total)
	}

	overview := map[string]any{
		"applicationStatus": "UP", "uptimeSeconds": int64(current.Sub(applicationStartedAt).Seconds()),
		"databaseStatus": databaseStatus["status"], "redisStatus": redisStatus["status"],
		"queueStatus": map[string]any{"status": "UP", "pendingJobs": pendingJobs},
		"diskUsage":   disk, "memoryUsage": map[string]any{"allocatedBytes": memory.Alloc, "systemBytes": memory.Sys, "heapObjects": memory.HeapObjects},
		"cpuUsage": map[string]any{"logicalCPUs": runtime.NumCPU()}, "goroutines": runtime.NumGoroutine(),
		"apiRequestRate": nil, "apiErrorRate": nil, "averageResponseTime": nil,
		"activeSessions": activeSessions, "scheduledTaskFailures": failedTasks,
		"paymentCallbackStatus": map[string]any{"totalLastHour": callback.Total, "successLastHour": callback.Success, "successRate": callbackRate},
	}
	return map[string]any{"overview": overview, "services": services, "providers": providers, "runtime": map[string]any{"goVersion": runtime.Version(), "goOS": runtime.GOOS, "goArch": runtime.GOARCH}}, nil
}

func (s *Service) MonitoringService(name string) (map[string]any, error) {
	data, err := s.Monitoring()
	if err != nil {
		return nil, err
	}
	result := map[string]any{"service": name, "sampledAt": now()}
	switch name {
	case "application":
		result["details"] = data["overview"]
	case "runtime":
		result["details"] = data["runtime"]
	case "providers":
		result["details"] = data["providers"]
	case "queue":
		result["details"] = data["overview"].(map[string]any)["queueStatus"]
	case "payment":
		result["details"] = data["overview"].(map[string]any)["paymentCallbackStatus"]
	case "database", "redis":
		for _, entry := range data["services"].([]map[string]any) {
			if entry["service"] == name {
				result["details"] = entry
				return result, nil
			}
		}
		return nil, NewError("UNKNOWN_SERVICE", "monitored service is unavailable")
	default:
		return nil, NewError("UNKNOWN_SERVICE", "unknown monitored service")
	}
	return result, nil
}
func statusOf(err error) string {
	if err == nil {
		return "UP"
	}
	return "DOWN"
}
func diskUsage(path string) map[string]any {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return map[string]any{"status": "UNAVAILABLE"}
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	percentage := float64(0)
	if total > 0 {
		percentage = float64(used) * 100 / float64(total)
	}
	return map[string]any{"totalBytes": total, "usedBytes": used, "freeBytes": free, "usedPercentage": percentage}
}
