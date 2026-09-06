/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// uniformCRLFContent has no mixed line endings, so the line-ending rule reports
// nothing about it and lint --fix leaves it alone. It is exactly the file the
// normalize command exists to repair.
const uniformCRLFContent = "KEY=value\r\nNEXT=2\r\n"

func TestNormalizeCommandRewritesFileThatLintFixLeavesAlone(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	mustWriteFile(t, path, uniformCRLFContent)

	// Regression guard for the finding-scoped --fix: no finding, no rewrite.
	if _, err := runLintCommand(t, root, "--fix"); err != nil {
		t.Fatalf("expected lint --fix to succeed, got %v", err)
	}
	if got := mustReadFile(t, path); got != uniformCRLFContent {
		t.Fatalf("lint --fix rewrote a file with no finding: got %q want %q", got, uniformCRLFContent)
	}

	// The normalize command is the opt-in surface that does rewrite it.
	stdout, stderr, err := runNormalizeCommandWithStreams(t, root)
	if err != nil {
		t.Fatalf("expected normalize to succeed, got %v", err)
	}
	if want := "normalized .env\n"; stdout != want {
		t.Fatalf("unexpected normalize report: got %q want %q", stdout, want)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	if got, want := mustReadFile(t, path), "KEY=value\nNEXT=2\n"; got != want {
		t.Fatalf("unexpected normalized content: got %q want %q", got, want)
	}
}

func TestNormalizeCommandLeavesNormalizedFileByteIdentical(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	const content = "KEY=value\n\nNEXT=2\n"
	mustWriteFile(t, path, content)

	stdout, stderr, err := runNormalizeCommandWithStreams(t, root)
	if err != nil {
		t.Fatalf("expected normalize to succeed, got %v", err)
	}
	// Nothing is reported because nothing was written.
	if stdout != "" {
		t.Fatalf("expected no normalize report, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	if got := mustReadFile(t, path); got != content {
		t.Fatalf("normalize rewrote an already normalized file: got %q want %q", got, content)
	}
}

func TestNormalizeCommandTargetNormalizesOnlyThatFile(t *testing.T) {
	root := t.TempDir()
	envPath := filepath.Join(root, ".env")
	stagingPath := filepath.Join(root, ".env.staging")
	mustWriteFile(t, envPath, uniformCRLFContent)
	mustWriteFile(t, stagingPath, uniformCRLFContent)

	stdout, _, err := runNormalizeCommandWithStreams(t, root, "--target=.env.staging")
	if err != nil {
		t.Fatalf("expected normalize --target to succeed, got %v", err)
	}
	if want := "normalized .env.staging\n"; stdout != want {
		t.Fatalf("unexpected normalize report: got %q want %q", stdout, want)
	}

	if got, want := mustReadFile(t, stagingPath), "KEY=value\nNEXT=2\n"; got != want {
		t.Fatalf("unexpected normalized content: got %q want %q", got, want)
	}
	if got := mustReadFile(t, envPath); got != uniformCRLFContent {
		t.Fatalf("normalize left the target scope: got %q want %q", got, uniformCRLFContent)
	}
}

func TestNormalizeCommandTargetDirNormalizesOnlyThatTree(t *testing.T) {
	root := t.TempDir()
	envPath := filepath.Join(root, ".env")
	nestedPath := filepath.Join(root, "src", "app", ".env.example")
	mustWriteFile(t, envPath, uniformCRLFContent)
	if err := os.MkdirAll(filepath.Dir(nestedPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	mustWriteFile(t, nestedPath, uniformCRLFContent)

	stdout, _, err := runNormalizeCommandWithStreams(t, root, "--target-dir=src")
	if err != nil {
		t.Fatalf("expected normalize --target-dir to succeed, got %v", err)
	}
	if want := "normalized " + filepath.Join("src", "app", ".env.example") + "\n"; stdout != want {
		t.Fatalf("unexpected normalize report: got %q want %q", stdout, want)
	}

	if got, want := mustReadFile(t, nestedPath), "KEY=value\nNEXT=2\n"; got != want {
		t.Fatalf("unexpected normalized content: got %q want %q", got, want)
	}
	if got := mustReadFile(t, envPath); got != uniformCRLFContent {
		t.Fatalf("normalize left the target-dir scope: got %q want %q", got, uniformCRLFContent)
	}
}

func TestNormalizeCommandRejectsTargetAndTargetDirTogether(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), uniformCRLFContent)

	_, _, err := runNormalizeCommandWithStreams(t, root, "--target=.env", "--target-dir=.")
	if err == nil {
		t.Fatal("expected normalize to reject both scope flags")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "--target and --target-dir cannot be used together") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeCommandRejectsPositionalArguments(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), uniformCRLFContent)

	if _, _, err := runNormalizeCommandWithStreams(t, root, ".env"); err == nil {
		t.Fatal("expected normalize to reject positional arguments")
	}
}

func runNormalizeCommandWithStreams(t *testing.T, root string, args ...string) (string, string, error) {
	t.Helper()

	withWorkingDir(t, root)

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(append([]string{"normalize"}, args...))

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s failed: %v", path, err)
	}
	return string(data)
}
