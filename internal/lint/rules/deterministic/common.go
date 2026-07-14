/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"regexp"

	"github.com/envaar/vaar/internal/lint"
)

// portableKeyPattern matches the key shape we treat as portable across
// shells, CI systems and dotenv parsers.
var portableKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// finding builds a Finding with the shared fields every rule supplies.
func finding(rule string, severity lint.Severity, file string, line int, message string) lint.Finding {
	return lint.Finding{
		Rule:     rule,
		Severity: severity,
		File:     file,
		Line:     line,
		Message:  message,
	}
}

// validKeyName accepts the portable env key format shared by the built-in
// rules and docs.
func validKeyName(key string) bool {
	return portableKeyPattern.MatchString(key)
}
