package lint

import "testing"

func TestMarkFixedFindingsIgnoresLineChangesForMatching(t *testing.T) {
	original := []Finding{{
		Rule:     "trailing-whitespace",
		Severity: SeverityWarn,
		File:     ".env",
		Line:     1,
		Message:  "line has trailing whitespace",
	}}

	remaining := []Finding{{
		Rule:     "trailing-whitespace",
		Severity: SeverityWarn,
		File:     ".env",
		Line:     2,
		Message:  "line has trailing whitespace",
	}}

	got := markFixedFindings(original, remaining)
	if len(got) != 1 {
		t.Fatalf("expected one retained finding, got %d", len(got))
	}

	if got[0].Fixed {
		t.Fatalf("expected remaining finding to stay present, but it was marked fixed")
	}

	if got[0].Line != 2 {
		t.Fatalf("expected remaining finding at line 2, got line %d", got[0].Line)
	}
}
