package analyzer

import (
	"fmt"
	"strings"

	"github.com/wadeling/origin-check/internal/probe"
)

const cacheSpeedRatioThreshold = 0.35

func scoreCache(r *probe.CacheProbeResult) (float64, string) {
	if r == nil || len(r.Runs) == 0 {
		return 50, "cache probe not run"
	}

	ok := r.SuccessfulRuns()
	if len(ok) == 0 {
		return 30, "cache probe failed on all runs"
	}

	score := 100.0
	var notes []string

	hashes := make(map[string]struct{})
	for _, run := range ok {
		hashes[run.ContentHash] = struct{}{}
	}
	sameBody := len(hashes) == 1 && len(ok) >= 2
	if sameBody && len(ok) >= 2 {
		score -= 25
		notes = append(notes, "identical response body across runs (possible cache reuse)")
	}

	if len(ok) >= 2 && ok[0].LatencyMS > 50 {
		ratio := float64(ok[1].LatencyMS) / float64(ok[0].LatencyMS)
		if ratio < cacheSpeedRatioThreshold {
			score -= 25
			notes = append(notes, fmt.Sprintf("2nd run latency %.0f%% of 1st (possible cache hit)", ratio*100))
		}
	}

	for _, run := range ok {
		if probe.CacheHeaderIndicatesHit(run.CacheHeaders) {
			score -= 30
			notes = append(notes, "CDN cache HIT in response headers")
			break
		}
	}

	if len(notes) == 0 {
		notes = append(notes, "no cache reuse indicators detected")
	}

	detail := fmt.Sprintf("%d/%d runs ok; %s", len(ok), len(r.Runs), strings.Join(notes, "; "))
	return clampScore(score), detail
}

func clampScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func cacheEvidenceDetail(r *probe.CacheProbeResult) string {
	if r == nil {
		return ""
	}
	var parts []string
	for _, run := range r.Runs {
		if run.Error != "" {
			parts = append(parts, fmt.Sprintf("#%d err=%s", run.Index, run.Error))
			continue
		}
		parts = append(parts, fmt.Sprintf("#%d %dms hash=%s headers=%s",
			run.Index, run.LatencyMS, run.ContentHash, run.CacheHeaders))
	}
	return strings.Join(parts, " | ")
}
