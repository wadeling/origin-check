package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/wadeling/origin-check/internal/probe"
)

const (
	alertMetadataMissing = "metadata_missing"
	alertMetadataPartial = "metadata_partial"
)

type metadataSample struct {
	Source string
	Model  string
}

func collectMetadataSamples(promptResults []PromptResult, cache *probe.CacheProbeResult) []metadataSample {
	var samples []metadataSample

	if cache != nil {
		for _, run := range cache.Runs {
			if run.Error != "" {
				continue
			}
			samples = append(samples, metadataSample{
				Source: fmt.Sprintf("cache#%d", run.Index),
				Model:  strings.TrimSpace(run.ResponseModel),
			})
		}
	}

	for _, pr := range promptResults {
		if pr.Response == nil || pr.Response.Error != "" {
			continue
		}
		source := pr.Case.Name
		if source == "" {
			source = "prompt"
		}
		samples = append(samples, metadataSample{
			Source: source,
			Model:  strings.TrimSpace(pr.Response.ResponseModel),
		})
	}

	return samples
}

type metadataEvaluation struct {
	Score  float64
	Detail string
	Alert  string
	Log    string
}

func evaluateMetadata(claimed string, samples []metadataSample, promptResults []PromptResult) metadataEvaluation {
	if len(samples) == 0 {
		return metadataEvaluation{
			Score:  25,
			Detail: "no successful probe responses to inspect API metadata",
			Alert:  alertMetadataMissing,
		}
	}

	var logLines []string
	withModel := make([]string, 0, len(samples))
	missing := 0
	for _, s := range samples {
		if s.Model == "" {
			missing++
			logLines = append(logLines, fmt.Sprintf("%s: (missing)", s.Source))
		} else {
			withModel = append(withModel, s.Model)
			logLines = append(logLines, fmt.Sprintf("%s: %s", s.Source, s.Model))
		}
	}

	total := len(samples)
	log := strings.Join(logLines, "\n")

	if len(withModel) == 0 {
		return metadataEvaluation{
			Score: 25,
			Detail: fmt.Sprintf(
				"0/%d requests returned the API model field; official OpenAI-compatible APIs always include model in JSON metadata",
				total,
			),
			Alert: alertMetadataMissing,
			Log:   log,
		}
	}

	unique := uniqueNormalizedModels(withModel)
	matchScore := scoreModelsAgainstClaimed(claimed, unique)

	score := matchScore
	alert := ""
	if missing > 0 {
		alert = alertMetadataPartial
		penalty := float64(missing) / float64(total) * 35
		score -= penalty
		if score < 0 {
			score = 0
		}
	}

	detail := fmt.Sprintf(
		"claimed=%s; metadata models=%s; %d/%d requests had model field",
		claimed,
		strings.Join(unique, ", "),
		len(withModel),
		total,
	)

	result := metadataEvaluation{
		Score:  round(score),
		Detail: detail,
		Alert:  alert,
		Log:    log,
	}

	if pr := findPromptResult(promptResults, "model_self_id"); pr != nil && pr.Response != nil {
		crossScore, crossAlert, crossDetail := scoreMetadataSelfReportCrossCheck(
			claimed,
			pr.Response.ResponseModel,
			pr.Response.Content,
		)
		if crossAlert != "" {
			result.Alert = crossAlert
			result.Detail = detail + "; " + crossDetail
			if crossScore < result.Score {
				result.Score = round(crossScore)
			}
		}
	}

	return result
}

func uniqueNormalizedModels(models []string) []string {
	seen := make(map[string]string)
	for _, m := range models {
		n := normalizeModel(m)
		if n == "" {
			continue
		}
		if _, ok := seen[n]; !ok {
			seen[n] = m
		}
	}
	out := make([]string, 0, len(seen))
	for _, m := range seen {
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

func scoreModelsAgainstClaimed(claimed string, unique []string) float64 {
	if len(unique) == 0 {
		return 25
	}
	if len(unique) > 1 {
		return 30
	}
	return scoreMetadata(claimed, unique[0])
}
