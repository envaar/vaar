/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package lint_test measures the runner against a representative dotenv
// fixture.
package lint_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules"
)

func BenchmarkRunnerSmoke(b *testing.B) {
	root := b.TempDir()
	fixture := []byte(strings.Join([]string{
		"KEY=value",
		"SPACED = value",
		"TRAILING=value  ",
		"SECOND=value",
		"",
	}, "\n"))

	if err := os.WriteFile(filepath.Join(root, ".env"), fixture, 0o644); err != nil {
		b.Fatalf("write benchmark fixture: %v", err)
	}

	runner := lint.NewRunner(rules.All()...)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := runner.Run(context.Background(), lint.Options{Root: root}); err != nil {
			b.Fatalf("benchmark run failed: %v", err)
		}
	}
}
