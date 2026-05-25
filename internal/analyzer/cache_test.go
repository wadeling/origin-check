package analyzer

import (
	"testing"

	"github.com/wadeling/origin-check/internal/probe"
)

func TestScoreCacheClean(t *testing.T) {
	score, detail := scoreCache(&probe.CacheProbeResult{
		Runs: []probe.CacheRun{
			{Index: 1, LatencyMS: 800, ContentHash: "aaa"},
			{Index: 2, LatencyMS: 750, ContentHash: "bbb"},
			{Index: 3, LatencyMS: 820, ContentHash: "ccc"},
		},
	})
	if score < 90 {
		t.Fatalf("expected high cache score, got %v detail=%s", score, detail)
	}
}

func TestScoreCacheSuspicious(t *testing.T) {
	score, _ := scoreCache(&probe.CacheProbeResult{
		Runs: []probe.CacheRun{
			{Index: 1, LatencyMS: 1000, ContentHash: "same", CacheHeaders: "cf-cache-status=HIT"},
			{Index: 2, LatencyMS: 200, ContentHash: "same"},
			{Index: 3, LatencyMS: 180, ContentHash: "same"},
		},
	})
	if score > 50 {
		t.Fatalf("expected low cache score for cache hit pattern, got %v", score)
	}
}

func TestScoreTraitsClaimedModel(t *testing.T) {
	traits := ExpectedTraits{MustContainClaimedModel: true}
	if score := scoreTraits("gpt-5.5", traits, "gpt-5.5"); score < 95 {
		t.Fatalf("expected match, got %v", score)
	}
	if score := scoreTraits("I am a helpful assistant", traits, "claude-opus-4-7"); score > 80 {
		t.Fatalf("expected penalty for mismatch, got %v", score)
	}
}
