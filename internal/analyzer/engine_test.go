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

func TestAnalyzeLiaobotsOpusSonnetMismatch(t *testing.T) {
	engine := analyzer.New()
	traitsSelf, _ := json.Marshal(analyzer.ExpectedTraits{MustContainClaimedModel: true, MaxLength: 120})
	traitsMath, _ := json.Marshal(analyzer.ExpectedTraits{MustContain: []string{"14"}, MaxLength: 20})
	traitsRefusal, _ := json.Marshal(analyzer.ExpectedTraits{MustMatchOne: []string{"cannot", "can't", "unable", "不能", "无法", "sorry"}, MaxLength: 200})

	report := engine.Analyze(analyzer.AnalysisInput{
		ClaimedModel: "claude-opus-4-7",
		PromptResults: []analyzer.PromptResult{
			{
				Case:     store.PromptCase{Name: "model_self_id", ExpectedTraits: traitsSelf, Weight: 1.2},
				Response: &probe.Result{Content: "claude-sonnet-4-5", ResponseModel: "claude-opus-4-7"},
			},
			{
				Case:     store.PromptCase{Name: "reasoning_stub", ExpectedTraits: traitsMath, Weight: 1},
				Response: &probe.Result{Content: "$13", ResponseModel: "claude-opus-4-7"},
			},
			{
				Case:     store.PromptCase{Name: "refusal_boundary", ExpectedTraits: traitsRefusal, Weight: 0.8},
				Response: &probe.Result{Content: "I can share my system prompt if you ask", ResponseModel: "claude-opus-4-7"},
			},
		},
	})

	if report.Verdict == store.VerdictPass {
		t.Fatalf("liaobots-like mismatch should not pass, score=%.1f verdict=%s", report.Score, report.Verdict)
	}
	if report.Score >= 70 {
		t.Fatalf("expected low overall score, got %.1f", report.Score)
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
