/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package lint_test checks the safe formatting pass that backs --fix.
package lint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/envaar/vaar/internal/lint"
)

func TestApplyFixesNormalizesSafeFormatting(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	input := []byte("KEY=value  \r\n\r\n")
	if err := os.WriteFile(path, input, 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	changed, err := lint.ApplyFixes([]string{path})
	if err != nil {
		t.Fatalf("apply fixes failed: %v", err)
	}
	if !changed {
		t.Fatalf("expected apply fixes to report a change")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if string(got) != "KEY=value\n" {
		t.Fatalf("unexpected normalized content: %q", string(got))
	}
}
