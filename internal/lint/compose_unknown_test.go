/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint_test

import (
	"bytes"
	"testing"

	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules"
)

// markerStripRule is an external-style rule with a fix half that is not named
// in the canonical fixOrder, standing in for a third-party FixableRule. Its Fix
// is order-sensitive: it resolves the X marker to "lf" or "crlf" depending on
// whether the data it receives still contains a CR, so the composition can
// prove not just that the fix ran but that it ran after the canonical prefix.
type markerStripRule struct{}

func (markerStripRule) ID() string                               { return "zz-custom" }
func (markerStripRule) Description() string                      { return "resolves the X marker by line-ending state" }
func (markerStripRule) Run(lint.Context) ([]lint.Finding, error) { return nil, nil }
func (markerStripRule) Fix(data []byte) []byte {
	if bytes.Contains(data, []byte("\r")) {
		return bytes.ReplaceAll(data, []byte("X"), []byte("crlf"))
	}
	return bytes.ReplaceAll(data, []byte("X"), []byte("lf"))
}

// replaceByteRule is an order-observable external rule: it rewrites one byte to
// another, so composing two of them in different orders yields different bytes.
type replaceByteRule struct {
	id       string
	from, to byte
}

func (r replaceByteRule) ID() string                             { return r.id }
func (replaceByteRule) Description() string                      { return "rewrites one byte to another" }
func (replaceByteRule) Run(lint.Context) ([]lint.Finding, error) { return nil, nil }
func (r replaceByteRule) Fix(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte{r.from}, []byte{r.to})
}

// TestFixDataAppliesUnknownFixableRule proves a fixable rule not named in the
// canonical fixOrder is still discovered as fixable and its fix half runs after
// the canonical prefix, rather than silently doing nothing. The assertion is
// order-sensitive, not just presence-sensitive: the input carries CRLF endings,
// and the custom fix resolves its marker to "lf" only if the CR is already gone
// by the time it runs. If the external fix ran before line-ending's Fix (the
// canonical fixOrder prefix), the CR would still be present and the marker
// would resolve to "crlf" instead, failing the "lf" assertion below.
func TestFixDataAppliesUnknownFixableRule(t *testing.T) {
	custom := markerStripRule{}

	if !lint.IsFixable(custom) {
		t.Fatalf("markerStripRule must read as fixable")
	}

	all := append(rules.All(), custom)
	// The canonical fixes normalize the CRLF ending and trim the trailing
	// whitespace before the unknown rule's fix resolves the marker; only then
	// has the CR already been normalized away.
	input := []byte("KEY=valueX  \r\n")
	want := []byte("KEY=valuelf\n")
	if got := lint.FixData(all, input); !bytes.Equal(got, want) {
		t.Fatalf("unknown fixable rule not applied after canonical prefix: got %q want %q", got, want)
	}
}

// TestFixDataAppliesUnknownFixesInIDOrder proves two fixable rules absent from
// fixOrder run in deterministic ascending-ID order regardless of the order they
// are passed in. Applied zz-a before zz-b, "1" -> "2" -> "3"; the reverse order
// would yield "2", so the asserted "3" pins the ID ordering.
func TestFixDataAppliesUnknownFixesInIDOrder(t *testing.T) {
	ruleA := replaceByteRule{id: "zz-a", from: '1', to: '2'}
	ruleB := replaceByteRule{id: "zz-b", from: '2', to: '3'}

	// Passed in reverse ID order to prove the composition sorts by ID.
	got := lint.FixData([]lint.Rule{ruleB, ruleA}, []byte("1\n"))
	if want := []byte("3\n"); !bytes.Equal(got, want) {
		t.Fatalf("unknown fixes ran out of ID order: got %q want %q", got, want)
	}
}
