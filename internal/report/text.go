/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package report

import (
	"fmt"
	"strings"

	"github.com/envaar/vaar/internal/lint"
)

// Text renders findings one per line in a stable, grep-friendly format and
// returns an empty string for clean runs.
func Text(findings []lint.Finding) string {
	if len(findings) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, finding := range findings {
		if i > 0 {
			builder.WriteByte('\n')
		}
		if finding.Fixed {
			builder.WriteString("[fixed] ")
		}
		builder.WriteString(fmt.Sprintf("%s %s %s:%d %s", finding.Severity, finding.Rule, finding.File, finding.Line, finding.Message))
	}
	builder.WriteByte('\n')
	return builder.String()
}
