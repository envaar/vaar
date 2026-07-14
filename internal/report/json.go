/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package report renders lint findings for humans and machines.
package report

import (
	"encoding/json"

	"github.com/envaar/vaar/internal/lint"
)

type payload struct {
	Findings []lint.Finding `json:"findings"`
}

// JSON renders findings as indented machine-readable output with an explicit
// findings array, even when the run found nothing.
func JSON(findings []lint.Finding) ([]byte, error) {
	if findings == nil {
		findings = []lint.Finding{}
	}

	return json.MarshalIndent(payload{Findings: findings}, "", "  ")
}
