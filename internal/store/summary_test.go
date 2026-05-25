package store

import "testing"

func TestSummarizeAuthenticityReportsSuspiciousTrend(t *testing.T) {
	reports := []AuthenticityReport{
		{Score: 55, Verdict: VerdictSuspicious},
		{Score: 52, Verdict: VerdictSuspicious},
		{Score: 58, Verdict: VerdictSuspicious},
		{Score: 100, Verdict: VerdictPass},
	}
	sum := SummarizeAuthenticityReports(reports)
	if sum == nil {
		t.Fatal("expected summary")
	}
	if sum.Verdict != VerdictSuspicious {
		t.Fatalf("expected suspicious verdict, got %s", sum.Verdict)
	}
	if sum.Score >= 90 {
		t.Fatalf("expected blended score below 90, got %v", sum.Score)
	}
}

func TestSummarizeAuthenticityReportsPassWhenConsistent(t *testing.T) {
	reports := []AuthenticityReport{
		{Score: 88, Verdict: VerdictPass},
		{Score: 85, Verdict: VerdictPass},
		{Score: 90, Verdict: VerdictPass},
	}
	sum := SummarizeAuthenticityReports(reports)
	if sum == nil {
		t.Fatal("expected summary")
	}
	if sum.Verdict != VerdictPass {
		t.Fatalf("expected pass, got %s", sum.Verdict)
	}
}
