package analyzer

import (
	"regexp"
	"strings"
)

var modelMentionPattern = regexp.MustCompile(`(?i)\b(gpt[-\d][\w.-]*|o\d[-\w]*|claude[-\w][\w.-]*|gemini[-\w][\w.-]*|deepseek[-\w][\w.-]*)\b`)

var claudeTiers = []string{"opus", "sonnet", "haiku"}

// scoreSelfReportedModel scores model_self_id responses against the claimed model ID.
func scoreSelfReportedModel(claimed, content string) float64 {
	claimed = normalizeModel(claimed)
	content = normalizeModel(content)
	if claimed == "" || content == "" {
		return 0
	}

	if content == claimed || strings.Contains(content, claimed) {
		return 100
	}

	mentions := extractModelMentions(content)
	if len(mentions) == 0 {
		return 40
	}

	worst := 100.0
	for _, m := range mentions {
		if modelsConflict(claimed, m) {
			return 10
		}
		s := scoreMetadata(claimed, m)
		if s < worst {
			worst = s
		}
	}
	return worst
}

// scoreMetadataSelfReportCrossCheck penalizes when API model field matches claimed but body self-reports another model.
func scoreMetadataSelfReportCrossCheck(claimed, apiModel, selfReport string) (float64, string, string) {
	claimed = normalizeModel(claimed)
	apiModel = normalizeModel(apiModel)
	selfReport = strings.TrimSpace(selfReport)
	if claimed == "" || apiModel == "" || selfReport == "" {
		return 100, "", ""
	}

	if !metadataMatchesClaimed(claimed, apiModel) {
		return 100, "", ""
	}

	mentions := extractModelMentions(normalizeModel(selfReport))
	for _, m := range mentions {
		if modelsConflict(claimed, m) {
			return 25, alertMetadataSelfReportMismatch,
				"API metadata=" + apiModel + " but self-report mentions " + m + " (tier/family mismatch)"
		}
	}

	if scoreSelfReportedModel(claimed, selfReport) < 50 {
		return 35, alertMetadataSelfReportMismatch,
			"API metadata matches claimed model but self-report identity is inconsistent"
	}

	return 100, "", ""
}

const alertMetadataSelfReportMismatch = "metadata_self_report_mismatch"

func extractModelMentions(text string) []string {
	text = normalizeModel(text)
	matches := modelMentionPattern.FindAllString(text, -1)
	seen := make(map[string]struct{})
	var out []string
	for _, m := range matches {
		m = normalizeModel(m)
		if m == "" {
			continue
		}
		if _, ok := seen[m]; !ok {
			seen[m] = struct{}{}
			out = append(out, m)
		}
	}
	return out
}

func modelsConflict(claimed, other string) bool {
	claimed = normalizeModel(claimed)
	other = normalizeModel(other)
	if claimed == "" || other == "" {
		return false
	}
	if claimed == other {
		return false
	}
	if strings.Contains(claimed, other) || strings.Contains(other, claimed) {
		return false
	}

	if strings.Contains(claimed, "claude") && strings.Contains(other, "claude") {
		cTier, oTier := claudeTier(claimed), claudeTier(other)
		if cTier != "" && oTier != "" && cTier != oTier {
			return true
		}
	}

	if strings.Contains(claimed, "gpt") && strings.Contains(other, "gpt") {
		if gptMajor(claimed) != gptMajor(other) && gptMajor(claimed) != "" && gptMajor(other) != "" {
			return true
		}
	}

	if strings.Contains(claimed, "gemini") && strings.Contains(other, "gemini") {
		if geminiMajor(claimed) != geminiMajor(other) && geminiMajor(claimed) != "" && geminiMajor(other) != "" {
			return true
		}
	}

	return false
}

func claudeTier(model string) string {
	for _, t := range claudeTiers {
		if strings.Contains(model, t) {
			return t
		}
	}
	return ""
}

func gptMajor(model string) string {
	for _, v := range []string{"5", "4", "3"} {
		if strings.Contains(model, "gpt-"+v) || strings.Contains(model, "gpt"+v) {
			return v
		}
	}
	return ""
}

func geminiMajor(model string) string {
	re := regexp.MustCompile(`gemini-(\d+)`)
	if m := re.FindStringSubmatch(model); len(m) == 2 {
		return m[1]
	}
	return ""
}

func metadataMatchesClaimed(claimed, apiModel string) bool {
	return scoreMetadata(claimed, apiModel) >= 80
}

func findPromptResult(results []PromptResult, name string) *PromptResult {
	for i := range results {
		if results[i].Case.Name == name && results[i].Response != nil && results[i].Response.Error == "" {
			return &results[i]
		}
	}
	return nil
}
