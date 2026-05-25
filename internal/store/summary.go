package store

const (
	AuthSummaryLookback = 6
	AuthSummaryRecent   = 3
)

type AuthenticitySummary struct {
	Score       float64
	Verdict     Verdict
	ReportCount int
}

func SummarizeAuthenticityReports(reports []AuthenticityReport) *AuthenticitySummary {
	if len(reports) == 0 {
		return nil
	}

	forScore := reports
	if len(forScore) > AuthSummaryLookback {
		forScore = forScore[:AuthSummaryLookback]
	}

	var sum float64
	for _, r := range forScore {
		sum += r.Score
	}
	avg := roundScore(sum / float64(len(forScore)))

	recent := reports
	if len(recent) > AuthSummaryRecent {
		recent = recent[:AuthSummaryRecent]
	}

	verdict := worstVerdict(recent)
	verdict = reconcileAuthVerdict(verdict, avg, recent)

	return &AuthenticitySummary{
		Score:       avg,
		Verdict:     verdict,
		ReportCount: len(forScore),
	}
}

func worstVerdict(reports []AuthenticityReport) Verdict {
	worst := VerdictUnknown
	rank := map[Verdict]int{
		VerdictUnknown:    0,
		VerdictPass:       1,
		VerdictSuspicious: 2,
		VerdictFail:       3,
	}
	for _, r := range reports {
		if rank[r.Verdict] > rank[worst] {
			worst = r.Verdict
		}
	}
	return worst
}

func reconcileAuthVerdict(trend Verdict, avg float64, recent []AuthenticityReport) Verdict {
	if len(recent) == 0 {
		return trend
	}

	switch {
	case avg < 50:
		return VerdictFail
	case avg < 75:
		if trend == VerdictPass || trend == VerdictUnknown {
			return VerdictSuspicious
		}
		return trend
	}

	if trend == VerdictFail {
		return VerdictFail
	}
	if trend == VerdictSuspicious {
		return VerdictSuspicious
	}
	if recent[0].Verdict != VerdictPass {
		return VerdictSuspicious
	}
	return VerdictPass
}

func roundScore(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
