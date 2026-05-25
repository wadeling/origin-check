package analyzer

import "testing"

func TestScoreSelfReportedModelOpusVsSonnet(t *testing.T) {
	score := scoreSelfReportedModel("claude-opus-4-7", "claude-sonnet-4-5")
	if score > 20 {
		t.Fatalf("expected low score for opus claimed vs sonnet self-report, got %v", score)
	}
}

func TestScoreSelfReportedModelExactMatch(t *testing.T) {
	score := scoreSelfReportedModel("claude-opus-4-7", "claude-opus-4-7")
	if score < 95 {
		t.Fatalf("expected high score, got %v", score)
	}
}

func TestModelsConflictOpusSonnet(t *testing.T) {
	if !modelsConflict("claude-opus-4-7", "claude-sonnet-4-5") {
		t.Fatal("expected opus vs sonnet conflict")
	}
}

func TestMetadataSelfReportCrossCheck(t *testing.T) {
	score, alert, detail := scoreMetadataSelfReportCrossCheck(
		"claude-opus-4-7",
		"claude-opus-4-7",
		"claude-sonnet-4-5",
	)
	if alert != alertMetadataSelfReportMismatch {
		t.Fatalf("expected mismatch alert, got %q detail=%s", alert, detail)
	}
	if score > 40 {
		t.Fatalf("expected low cross-check score, got %v", score)
	}
}

func TestReasoningStubWrongAnswer(t *testing.T) {
	traits := ExpectedTraits{MustContain: []string{"14"}, MaxLength: 20}
	score := scoreTraits("$13", traits, "claude-opus-4-7")
	if score >= 80 {
		t.Fatalf("expected low score for wrong math answer, got %v", score)
	}
}

func TestScoreTraitsSelfIDUsesStrictModelMatch(t *testing.T) {
	traits := ExpectedTraits{MustContainClaimedModel: true, MaxLength: 120}
	score := scoreTraits("claude-sonnet-4-5", traits, "claude-opus-4-7")
	if score > 20 {
		t.Fatalf("expected strict self-id failure, got %v", score)
	}
}
