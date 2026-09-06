/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules"
)

// uniformCRLF has no mixed line endings, so the line-ending rule reports
// nothing and the finding-scoped --fix pipeline must leave it alone. It is the
// exact input the normalize pass exists to repair.
const uniformCRLF = "KEY=value\r\nNEXT=2\r\n"

func TestNormalizeDataForcesLineEndingsThatFixDataKeeps(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"uniform-crlf", uniformCRLF, "KEY=value\nNEXT=2\n"},
		{"lone-cr", "A=1\rB=2\r", "A=1\nB=2\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The rule fix halves behind --fix are finding-scoped: uniform
			// endings produce no finding, so composing every rule changes
			// nothing here.
			if got := string(lint.FixData(rules.All(), []byte(tc.input))); got != tc.input {
				t.Fatalf("--fix pipeline rewrote a file with no finding: got %q want %q", got, tc.input)
			}

			// NormalizeData is deliberately blunt and rewrites the endings anyway.
			if got := string(lint.NormalizeData([]byte(tc.input))); got != tc.want {
				t.Fatalf("unexpected normalized bytes: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeDataAppliesTheWholeFilePipeline(t *testing.T) {
	input := append([]byte{0xEF, 0xBB, 0xBF}, []byte("KEY=value  \r\n\r\n\r\nFOO=bar   ")...)

	// BOM stripped, endings forced to LF, trailing whitespace trimmed, the
	// blank run collapsed to one and exactly one final newline left behind.
	if got, want := string(lint.NormalizeData(input)), "KEY=value\n\nFOO=bar\n"; got != want {
		t.Fatalf("unexpected normalized bytes: got %q want %q", got, want)
	}
}

func TestNormalizeRewritesUniformCRLFFileAndReportsIt(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte(uniformCRLF), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	normalized, err := lint.Normalize(lint.Options{Root: root})
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	if len(normalized) != 1 || normalized[0] != ".env" {
		t.Fatalf("unexpected normalized paths: %q", normalized)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if want := "KEY=value\nNEXT=2\n"; string(got) != want {
		t.Fatalf("unexpected normalized content: got %q want %q", string(got), want)
	}
}

func TestNormalizeLeavesAlreadyNormalizedFileUntouched(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	const content = "KEY=value\n\nNEXT=2\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	normalized, err := lint.Normalize(lint.Options{Root: root})
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	// An empty report means the file was skipped rather than rewritten.
	if len(normalized) != 0 {
		t.Fatalf("expected no file to be rewritten, got %q", normalized)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != content {
		t.Fatalf("normalized file changed: got %q want %q", string(got), content)
	}
}

func TestNormalizeIsStableAcrossRuns(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte(uniformCRLF), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	first, err := lint.Normalize(lint.Options{Root: root})
	if err != nil {
		t.Fatalf("first normalize failed: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected the first run to rewrite the file, got %q", first)
	}

	second, err := lint.Normalize(lint.Options{Root: root})
	if err != nil {
		t.Fatalf("second normalize failed: %v", err)
	}
	if len(second) != 0 {
		t.Fatalf("expected the second run to rewrite nothing, got %q", second)
	}
}
