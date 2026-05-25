package analyzer_test

import (
	"encoding/json"
	"testing"

	"github.com/wadeling/origin-check/internal/analyzer"
	"github.com/wadeling/origin-check/internal/probe"
	"github.com/wadeling/origin-check/internal/store"
)

func TestAnalyzePassWhenMetadataAndPromptsMatch(t *testing.T) {
	engine := analyzer.New()
	traits, _ := json.Marshal(map[string]interface{}{
		"must_contain": []string{"391"},
		"max_length":   10,
	})

	report := engine.Analyze(analyzer.AnalysisInput{
		ClaimedModel: "gpt-5.5",
		PromptResults: []analyzer.PromptResult{
			{
				Case: store.PromptCase{
					Name:           "math",
					ExpectedTraits: traits,
					Weight:         1,
				},
				Response: &probe.Result{Content: "391", ResponseModel: "gpt-5.5"},
			},
		},
		CacheResult: &probe.CacheProbeResult{
			Runs: []probe.CacheRun{
				{LatencyMS: 500, ContentHash: "a"},
				{LatencyMS: 480, ContentHash: "b"},
			},
		},
	})

	if report.Score < 70 {
		t.Fatalf("expected pass-level score, got %v", report.Score)
	}
	if report.Verdict != store.VerdictPass && report.Verdict != store.VerdictSuspicious {
		t.Fatalf("unexpected verdict: %s", report.Verdict)
	}
}

func TestAnalyzeFailWhenMetadataMismatch(t *testing.T) {
	engine := analyzer.New()
	report := engine.Analyze(analyzer.AnalysisInput{
		ClaimedModel: "gpt-5.5",
		PromptResults: []analyzer.PromptResult{
			{Response: &probe.Result{ResponseModel: "gpt-3.5-turbo"}},
		},
	})

	if report.Score > 60 {
		t.Fatalf("expected lower score for mismatch, got %v", report.Score)
	}
}
