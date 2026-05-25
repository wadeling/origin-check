package probe

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// CacheProbePrompt is a fixed prompt for repeatability / CDN cache detection.
const CacheProbePrompt = "CACHE_PROBE_v1_b9e4d2"

const DefaultCacheRepeats = 3

type CacheRun struct {
	Index         int
	LatencyMS     int
	ContentHash   string
	ResponseModel string
	CacheHeaders  string
	Error         string
}

type CacheProbeResult struct {
	Runs []CacheRun
}

// RunCacheProbe sends identical chat requests (temperature 0) and records latency + body hash.
func (c *Client) RunCacheProbe(ctx context.Context, ep Endpoint, model string, repeats int) *CacheProbeResult {
	if repeats <= 0 {
		repeats = DefaultCacheRepeats
	}
	temp := 0.0
	maxTok := 64
	out := &CacheProbeResult{Runs: make([]CacheRun, 0, repeats)}

	for i := 0; i < repeats; i++ {
		run := CacheRun{Index: i + 1}
		res, err := c.ChatCompletionOpts(ctx, ep, model, CacheProbePrompt, ChatOpts{
			Stream:      false,
			Temperature: &temp,
			MaxTokens:   &maxTok,
		})
		if err != nil {
			run.Error = err.Error()
			out.Runs = append(out.Runs, run)
			continue
		}
		run.LatencyMS = res.LatencyMS
		run.ContentHash = res.ContentHash
		run.ResponseModel = res.ResponseModel
		run.CacheHeaders = res.CacheHeaders
		if res.Error != "" {
			run.Error = res.Error
		}
		out.Runs = append(out.Runs, run)
	}
	return out
}

var cacheHeaderKeys = []string{
	"cf-cache-status",
	"age",
	"cache-control",
	"x-cache",
	"x-cache-hits",
	"x-cache-status",
	"x-request-cache",
	"cdn-cache",
	"x-fastly-cache",
}

func FormatCacheHeaderHints(h http.Header) string {
	if h == nil {
		return ""
	}
	var parts []string
	seen := make(map[string]bool)
	for _, key := range cacheHeaderKeys {
		v := h.Get(key)
		if v == "" {
			continue
		}
		kl := strings.ToLower(key)
		if seen[kl] {
			continue
		}
		seen[kl] = true
		parts = append(parts, fmt.Sprintf("%s=%s", key, v))
	}
	return strings.Join(parts, "; ")
}

var (
	cfCacheHitRE = regexp.MustCompile(`(?i)cf-cache-status\s*[:=]\s*hit\b`)
	xCacheHitRE  = regexp.MustCompile(`(?i)x-cache[^;,\s]*\bhit\b`)
)

func CacheHeaderIndicatesHit(hints string) bool {
	if hints == "" {
		return false
	}
	return cfCacheHitRE.MatchString(hints) || xCacheHitRE.MatchString(hints)
}

func (r *CacheProbeResult) SuccessfulRuns() []CacheRun {
	var ok []CacheRun
	for _, run := range r.Runs {
		if run.Error == "" && run.ContentHash != "" {
			ok = append(ok, run)
		}
	}
	return ok
}
