package trigger_test

import (
	"context"
	"testing"

	"github.com/wadeling/origin-check/internal/store"
	"github.com/wadeling/origin-check/internal/trigger"
)

func TestModelsForJobAuthenticityAllClaimed(t *testing.T) {
	relay := store.Relay{
		ClaimedModels: []string{"gpt-5.5", "claude-opus-4-7"},
		HealthModel:   "gpt-5.4-mini",
	}
	models := trigger.ModelsForJob(relay, store.JobAuthenticity, "")
	if len(models) != 2 || models[0] != "gpt-5.5" {
		t.Fatalf("unexpected models: %v", models)
	}
}

func TestModelsForJobHealthUsesHealthModel(t *testing.T) {
	relay := store.Relay{
		ClaimedModels: []string{"gpt-5.5"},
		HealthModel:   "gpt-5.4-mini",
	}
	models := trigger.ModelsForJob(relay, store.JobHealth, "")
	if len(models) != 1 || models[0] != "gpt-5.4-mini" {
		t.Fatalf("unexpected models: %v", models)
	}
}

func TestModelsForJobSingleFilter(t *testing.T) {
	relay := store.Relay{ClaimedModels: []string{"gpt-5.5", "claude-opus-4-7"}}
	models := trigger.ModelsForJob(relay, store.JobAuthenticity, "claude-opus-4-7")
	if len(models) != 1 || models[0] != "claude-opus-4-7" {
		t.Fatalf("unexpected models: %v", models)
	}
}

func TestModelsForJobEmptyFilterIgnored(t *testing.T) {
	_ = context.Background()
	relay := store.Relay{HealthModel: "gpt-5.4-mini"}
	models := trigger.ModelsForJob(relay, store.JobPerformance, "")
	if len(models) != 1 {
		t.Fatalf("expected fallback health model, got %v", models)
	}
}
