package config

import (
	"fmt"
	"os"
	"time"
)

// ScheduleConfig controls probe job intervals for the scheduler.
type ScheduleConfig struct {
	HealthInterval        time.Duration
	PerformanceInterval   time.Duration
	AuthenticityInterval  time.Duration
	HealthOnStartup       bool
	PerformanceOnStartup  bool
	AuthenticityOnStartup bool
}

func LoadScheduleConfig() (ScheduleConfig, error) {
	health, err := durationEnv("PROBE_HEALTH_INTERVAL", 15*time.Minute)
	if err != nil {
		return ScheduleConfig{}, err
	}
	performance, err := durationEnv("PROBE_PERFORMANCE_INTERVAL", 24*time.Hour)
	if err != nil {
		return ScheduleConfig{}, err
	}
	authenticity, err := durationEnv("PROBE_AUTHENTICITY_INTERVAL", 24*time.Hour)
	if err != nil {
		return ScheduleConfig{}, err
	}

	cfg := ScheduleConfig{
		HealthInterval:        health,
		PerformanceInterval:   performance,
		AuthenticityInterval:  authenticity,
		HealthOnStartup:       boolEnv("PROBE_HEALTH_ON_STARTUP", true),
		PerformanceOnStartup:  boolEnv("PROBE_PERFORMANCE_ON_STARTUP", false),
		AuthenticityOnStartup: boolEnv("PROBE_AUTHENTICITY_ON_STARTUP", false),
	}

	if cfg.HealthInterval < time.Minute {
		return ScheduleConfig{}, fmt.Errorf("PROBE_HEALTH_INTERVAL must be at least 1m")
	}
	if cfg.PerformanceInterval < time.Minute {
		return ScheduleConfig{}, fmt.Errorf("PROBE_PERFORMANCE_INTERVAL must be at least 1m")
	}
	if cfg.AuthenticityInterval < time.Minute {
		return ScheduleConfig{}, fmt.Errorf("PROBE_AUTHENTICITY_INTERVAL must be at least 1m")
	}

	return cfg, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid duration %q: %w", key, v, err)
	}
	return d, nil
}

func boolEnv(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}
