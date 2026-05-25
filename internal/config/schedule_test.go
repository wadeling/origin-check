package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/wadeling/origin-check/internal/config"
)

func TestLoadScheduleConfigDefaults(t *testing.T) {
	os.Unsetenv("PROBE_HEALTH_INTERVAL")
	os.Unsetenv("PROBE_PERFORMANCE_INTERVAL")
	os.Unsetenv("PROBE_AUTHENTICITY_INTERVAL")
	os.Unsetenv("PROBE_AUTHENTICITY_ON_STARTUP")

	cfg, err := config.LoadScheduleConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HealthInterval != 15*time.Minute {
		t.Fatalf("health interval: got %v", cfg.HealthInterval)
	}
	if cfg.PerformanceInterval != 24*time.Hour {
		t.Fatalf("performance interval: got %v", cfg.PerformanceInterval)
	}
	if cfg.PerformanceOnStartup {
		t.Fatal("performance should not run on startup by default")
	}
	if cfg.AuthenticityInterval != 24*time.Hour {
		t.Fatalf("authenticity interval: got %v", cfg.AuthenticityInterval)
	}
	if cfg.AuthenticityOnStartup {
		t.Fatal("authenticity should not run on startup by default")
	}
}

func TestLoadScheduleConfigCustom(t *testing.T) {
	t.Setenv("PROBE_PERFORMANCE_INTERVAL", "12h")
	t.Setenv("PROBE_PERFORMANCE_ON_STARTUP", "true")

	cfg, err := config.LoadScheduleConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PerformanceInterval != 12*time.Hour {
		t.Fatalf("performance interval: got %v", cfg.PerformanceInterval)
	}
	if !cfg.PerformanceOnStartup {
		t.Fatal("expected performance on startup")
	}
}
