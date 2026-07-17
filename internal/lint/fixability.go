/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint

// FixableRule is an optional interface a Rule may implement to supply the fix
// half that repairs its own findings. It stays separate from Rule on purpose:
// widening Rule would force every external rule implementer to recompile, so
// the fix half is expressed as an opt-in marker discovered through a type
// assertion instead.
//
// Fixability is intrinsic and cannot drift: a rule is fixable exactly when it
// carries a Fix method, and IsFixable and the --list-rules FIXABLE column both
// key on the presence of that method rather than a hand-maintained claim.
// ApplyFixes composes these Fix halves in a canonical order to repair files.
type FixableRule interface {
	// Fix returns data with this rule's findings repaired, so composing fix
	// halves scopes --fix to selected rules. Repairing a finding may normalize
	// the related file-wide property (for example converting every line ending
	// to LF, or emptying whitespace-only lines), so bytes beyond the reported
	// finding are not guaranteed to be left untouched.
	Fix(data []byte) []byte
}

// IsFixable reports whether a rule supplies a fix half through FixableRule.
// Rules that do not implement the marker are treated as not auto-fixable.
func IsFixable(r Rule) bool {
	_, ok := r.(FixableRule)
	return ok
}
