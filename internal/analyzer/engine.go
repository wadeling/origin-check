package analyzer

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/wadeling/origin-check/internal/probe"
	"github.com/wadeling/origin-check/internal/store"
)

const (
	weightMetadata = 0.15
	weightPrompt   = 0.40
	weightCapability = 0.25
	weightBaseline = 0.20
)

type ExpectedTraits struct {
	MustContain             []string `json:"must_contain,omitempty"`
	MustNotContain          []string `json:"must_not_contain,omitempty"`
	MustMatchOne            []string `json:"must_match_one,omitempty"`
	MustContainClaimedModel bool     `json:"must_contain_claimed_model,omitempty"`
	MustBeJSON              bool     `json:"must_be_json,omitempty"`
	MaxLength               int      `json:"max_length,omitempty"`
	WordCountMin            int      `json:"word_count_min,omitempty"`
	WordCountMax            int      `json:"word_count_max,omitempty"`
}

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

type AnalysisInput struct {
	ClaimedModel  string
	PromptResults []PromptResult
	CacheResult   *probe.CacheProbeResult
}

type PromptResult struct {
	Case     store.PromptCase
	Response *probe.Result
}

func (e *Engine) Analyze(in AnalysisInput) store.AuthenticityReport {
	signals := make([]store.SignalEvidence, 0)
	totalWeight := 0.0
	weightedScore := 0.0

	// Metadata signal (aggregate model field from all probe requests)
	metaSamples := collectMetadataSamples(in.PromptResults, in.CacheResult)
	meta := evaluateMetadata(in.ClaimedModel, metaSamples)
	metaWeight := weightMetadata
	metaSignal := store.SignalEvidence{
		Signal: "metadata",
		Score:  meta.Score,
		Weight: metaWeight,
		Detail: meta.Detail,
		Alert:  meta.Alert,
	}
	if meta.Log != "" {
		metaSignal.Response = meta.Log
	}
	signals = append(signals, metaSignal)
	weightedScore += meta.Score * metaWeight
	totalWeight += metaWeight

	// Prompt fingerprint signals
	promptTotal := 0.0
	promptCount := 0
	for _, pr := range in.PromptResults {
		if pr.Response == nil || pr.Response.Error != "" {
			continue
		}
		var traits ExpectedTraits
		_ = json.Unmarshal(pr.Case.ExpectedTraits, &traits)
		score := scoreTraits(pr.Response.Content, traits, in.ClaimedModel)
		promptTotal += score
		promptCount++
		signals = append(signals, store.SignalEvidence{
			Signal:   "prompt:" + pr.Case.Name,
			Score:    score,
			Weight:   pr.Case.Weight,
			Detail:   traitDetail(score),
			Prompt:   probe.Summarize(pr.Case.Prompt, 80),
			Response: probe.Summarize(pr.Response.Content, 120),
		})
	}

	promptScore := 50.0
	if promptCount > 0 {
		promptScore = promptTotal / float64(promptCount)
	}
	signals = append(signals, store.SignalEvidence{
		Signal: "prompt_suite",
		Score:  promptScore,
		Weight: weightPrompt,
		Detail: "fingerprint prompt suite average",
	})
	weightedScore += promptScore * weightPrompt
	totalWeight += weightPrompt

	// Capability placeholder (MVP: assume pass if prompts mostly succeed)
	capScore := promptScore
	signals = append(signals, store.SignalEvidence{
		Signal: "capability",
		Score:  capScore,
		Weight: weightCapability,
		Detail: "basic capability inferred from prompt compliance",
	})
	weightedScore += capScore * weightCapability
	totalWeight += weightCapability

	cacheScore, cacheDetail := scoreCache(in.CacheResult)
	signals = append(signals, store.SignalEvidence{
		Signal:   "cache",
		Score:    cacheScore,
		Weight:   weightBaseline,
		Detail:   cacheDetail,
		Response: cacheEvidenceDetail(in.CacheResult),
	})
	weightedScore += cacheScore * weightBaseline
	totalWeight += weightBaseline

	finalScore := weightedScore / totalWeight
	confidence := float64(promptCount) / float64(max(len(in.PromptResults), 1)) * 100

	verdict := store.VerdictUnknown
	switch {
	case finalScore >= 75 && confidence >= 60:
		verdict = store.VerdictPass
	case finalScore >= 50:
		verdict = store.VerdictSuspicious
	default:
		verdict = store.VerdictFail
	}

	if meta.Alert == alertMetadataMissing && verdict == store.VerdictPass {
		verdict = store.VerdictSuspicious
	}

	return store.AuthenticityReport{
		ClaimedModel: in.ClaimedModel,
		Score:        round(finalScore),
		Confidence:   round(confidence),
		Verdict:      verdict,
		Signals:      signals,
	}
}

func scoreMetadata(claimed, response string) float64 {
	if response == "" {
		return 50
	}
	if normalizeModel(claimed) == normalizeModel(response) {
		return 100
	}
	if strings.Contains(normalizeModel(response), normalizeModel(claimed)) ||
		strings.Contains(normalizeModel(claimed), normalizeModel(response)) {
		return 80
	}
	return 30
}

func normalizeModel(m string) string {
	return strings.ToLower(strings.TrimSpace(m))
}

func scoreTraits(content string, t ExpectedTraits, claimedModel string) float64 {
	content = strings.TrimSpace(content)
	if content == "" {
		return 0
	}

	score := 100.0
	lower := strings.ToLower(content)

	if t.MustContainClaimedModel && claimedModel != "" {
		norm := normalizeModel(claimedModel)
		resp := normalizeModel(content)
		if !strings.Contains(resp, norm) && !strings.Contains(norm, resp) {
			// allow partial token match e.g. gpt-5.5 in longer self-id line
			parts := strings.FieldsFunc(norm, func(r rune) bool {
				return r == '-' || r == '.' || r == '_'
			})
			found := false
			for _, p := range parts {
				if len(p) >= 3 && strings.Contains(resp, p) {
					found = true
					break
				}
			}
			if !found {
				score -= 35
			}
		}
	}

	for _, s := range t.MustContain {
		if !strings.Contains(lower, strings.ToLower(s)) {
			score -= 25
		}
	}
	for _, s := range t.MustNotContain {
		if strings.Contains(content, s) {
			score -= 20
		}
	}
	if len(t.MustMatchOne) > 0 {
		found := false
		for _, s := range t.MustMatchOne {
			if strings.EqualFold(strings.TrimSpace(content), s) || strings.Contains(lower, strings.ToLower(s)) {
				found = true
				break
			}
		}
		if !found {
			score -= 30
		}
	}
	if t.MustBeJSON {
		var js json.RawMessage
		if err := json.Unmarshal([]byte(content), &js); err != nil {
			// try extract JSON from content
			re := regexp.MustCompile(`\{.*\}`)
			if match := re.FindString(content); match != "" {
				if err2 := json.Unmarshal([]byte(match), &js); err2 != nil {
					score -= 40
				}
			} else {
				score -= 40
			}
		}
	}
	if t.MaxLength > 0 && len(content) > t.MaxLength {
		score -= 15
	}
	if t.WordCountMin > 0 || t.WordCountMax > 0 {
		words := len(strings.Fields(content))
		if t.WordCountMin > 0 && words < t.WordCountMin {
			score -= 20
		}
		if t.WordCountMax > 0 && words > t.WordCountMax {
			score -= 15
		}
	}

	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func traitDetail(score float64) string {
	if score >= 80 {
		return "matches expected traits"
	}
	if score >= 50 {
		return "partially matches expected traits"
	}
	return "does not match expected traits"
}

func round(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
