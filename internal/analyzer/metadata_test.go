package analyzer

import (
	"testing"

	"github.com/wadeling/origin-check/internal/probe"
	"github.com/wadeling/origin-check/internal/store"
)

func TestEvaluateMetadataMissing(t *testing.T) {
	meta := evaluateMetadata("gpt-5.5", []metadataSample{
		{Source: "cache#1", Model: ""},
		{Source: "math_simple", Model: ""},
	}, nil)
	if meta.Alert != alertMetadataMissing {
		t.Fatalf("expected metadata_missing alert, got %q", meta.Alert)
	}
	if meta.Score > 30 {
		t.Fatalf("expected low score, got %v", meta.Score)
	}
}

func TestEvaluateMetadataFromAllRequests(t *testing.T) {
	meta := evaluateMetadata("gpt-5.5", []metadataSample{
		{Source: "cache#1", Model: "gpt-5.5"},
		{Source: "cache#2", Model: ""},
		{Source: "model_self_id", Model: "gpt-5.5"},
	}, nil)
	if meta.Alert != alertMetadataPartial {
		t.Fatalf("expected metadata_partial alert, got %q", meta.Alert)
	}
	if meta.Score < 70 {
		t.Fatalf("expected decent score with mostly matching models, got %v", meta.Score)
	}
}

func TestEvaluateMetadataInconsistentModels(t *testing.T) {
	meta := evaluateMetadata("gpt-5.5", []metadataSample{
		{Source: "a", Model: "gpt-5.5"},
		{Source: "b", Model: "gpt-4o"},
	}, nil)
	if meta.Score > 35 {
		t.Fatalf("expected low score for inconsistent metadata models, got %v", meta.Score)
	}
}

func TestEvaluateMetadataSelfReportMismatch(t *testing.T) {
	meta := evaluateMetadata("claude-opus-4-7", []metadataSample{
		{Source: "model_self_id", Model: "claude-opus-4-7"},
	}, []PromptResult{
		{
			Case:     store.PromptCase{Name: "model_self_id"},
			Response: &probe.Result{Content: "claude-sonnet-4-5", ResponseModel: "claude-opus-4-7"},
		},
	})
	if meta.Alert != alertMetadataSelfReportMismatch {
		t.Fatalf("expected self-report mismatch alert, got %q", meta.Alert)
	}
	if meta.Score > 40 {
		t.Fatalf("expected low metadata score, got %v", meta.Score)
	}
}

func TestAnalyzeMetadataMissingCapsVerdict(t *testing.T) {
	engine := New()
	report := engine.Analyze(AnalysisInput{
		ClaimedModel: "gpt-5.5",
		PromptResults: []PromptResult{
			{
				Case:     store.PromptCase{Name: "math", Weight: 1},
				Response: &probe.Result{Content: "391", ResponseModel: ""},
			},
		},
		CacheResult: &probe.CacheProbeResult{
			Runs: []probe.CacheRun{{Index: 1, ContentHash: "x"}},
		},
	})
	var metaAlert string
	for _, s := range report.Signals {
		if s.Signal == "metadata" {
			metaAlert = s.Alert
		}
	}
	if metaAlert != alertMetadataMissing {
		t.Fatalf("expected metadata_missing, got %q", metaAlert)
	}
	if report.Verdict == store.VerdictPass {
		t.Fatalf("metadata missing should not yield pass verdict, got %s", report.Verdict)
	}
}
