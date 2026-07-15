/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package cli exercises the lint command end to end, including exit codes,
// fixes, rendering and failure paths.
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/envaar/vaar/internal/lint"
)

func TestLintCommandUnknownRuleReturnsExitCode2(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"lint", "--only=json-output"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}

	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "unknown lint rule \"json-output\"") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLintCommandEmptySelectionReturnsExitCode2(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"lint", "--only=duplicate-key", "--skip=duplicate-key"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}

	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "no lint rules selected after applying --only and --skip") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLintCommandReportsFindingsInTextAndJSON(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("KEY=value\nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	withWorkingDir(t, root)

	t.Run("text", func(t *testing.T) {
		var stdout bytes.Buffer

		cmd := newRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetArgs([]string{"lint"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected an error")
		}
		if got := ExitCode(err); got != ExitFindings {
			t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
		}

		want := "error duplicate-key .env:2 KEY is defined more than once\n"
		if got := stdout.String(); got != want {
			t.Fatalf("unexpected output: got %q want %q", got, want)
		}
	})

	t.Run("json", func(t *testing.T) {
		var stdout bytes.Buffer

		cmd := newRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetArgs([]string{"lint", "--json"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected an error")
		}
		if got := ExitCode(err); got != ExitFindings {
			t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
		}

		var payload struct {
			Findings []lint.Finding `json:"findings"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if got, want := len(payload.Findings), 1; got != want {
			t.Fatalf("unexpected finding count: got %d want %d", got, want)
		}
		if got, want := payload.Findings[0].Rule, "duplicate-key"; got != want {
			t.Fatalf("unexpected rule: got %q want %q", got, want)
		}
		if got, want := payload.Findings[0].File, ".env"; got != want {
			t.Fatalf("unexpected file: got %q want %q", got, want)
		}
		if got, want := payload.Findings[0].Line, 2; got != want {
			t.Fatalf("unexpected line: got %d want %d", got, want)
		}
	})
}

func TestLintCommandRejectsOutputWithoutJSON(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("KEY=value\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--output=lint-report.json")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "--output requires --json") {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout output, got %q", stdout)
	}
}

func TestLintCommandWritesJSONOutputToFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("KEY=value\nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--json", "--output=lint-report.json")
	if err == nil {
		t.Fatal("expected findings error")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout output, got %q", stdout)
	}

	data, err := os.ReadFile(filepath.Join(root, "lint-report.json"))
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	var payload struct {
		Findings []lint.Finding `json:"findings"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got, want := len(payload.Findings), 1; got != want {
		t.Fatalf("unexpected finding count: got %d want %d", got, want)
	}
	if got, want := payload.Findings[0].Rule, "duplicate-key"; got != want {
		t.Fatalf("unexpected rule: got %q want %q", got, want)
	}
}

func TestLintCommandFixesSafeFormatting(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("KEY=value  \n\n\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--fix")
	if err != nil {
		t.Fatalf("expected lint --fix to succeed, got %v", err)
	}
	if stdout != "" {
		t.Fatalf("expected no output after fixing, got %q", stdout)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != "KEY=value\n" {
		t.Fatalf("unexpected fixed content: got %q want %q", string(got), "KEY=value\n")
	}
}

func TestLintCommandAcceptsRepeatableOnlyFlags(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("KEY=value  \nKEY=value-2\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--json", "--only=duplicate-key", "--only=trailing-whitespace")
	if err == nil {
		t.Fatal("expected findings error")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}

	var payload struct {
		Findings []lint.Finding `json:"findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got, want := len(payload.Findings), 2; got != want {
		t.Fatalf("unexpected finding count: got %d want %d", got, want)
	}

	rules := map[string]bool{}
	for _, finding := range payload.Findings {
		rules[finding.Rule] = true
	}
	if !rules["duplicate-key"] {
		t.Fatalf("expected duplicate-key finding, got %#v", payload.Findings)
	}
	if !rules["trailing-whitespace"] {
		t.Fatalf("expected trailing-whitespace finding, got %#v", payload.Findings)
	}
}

func TestLintCommandAcceptsRepeatableSkipFlags(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("KEY=value  \n\n\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--skip=trailing-whitespace", "--skip=extra-blank-line", "--skip=ending-blank-line")
	if err != nil {
		t.Fatalf("expected lint --skip to succeed, got %v", err)
	}
	if stdout != "" {
		t.Fatalf("expected no output when repeatable skip flags remove all findings, got %q", stdout)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != "KEY=value  \n\n\n" {
		t.Fatalf("unexpected file content after skip-only run: %q", string(got))
	}
}

func TestLintCommandTargetFileModeSupportsJsonAndOnly(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("ROOT=value\nROOT=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env.staging"), []byte("STAGING=value\nSTAGING=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--json", "--only=duplicate-key", "--target=.env.staging")
	if err == nil {
		t.Fatal("expected findings error")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}

	var payload struct {
		Findings []lint.Finding `json:"findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got, want := len(payload.Findings), 1; got != want {
		t.Fatalf("unexpected finding count: got %d want %d", got, want)
	}
	if got, want := payload.Findings[0].File, ".env.staging"; got != want {
		t.Fatalf("unexpected finding file: got %q want %q", got, want)
	}
	if got, want := payload.Findings[0].Rule, "duplicate-key"; got != want {
		t.Fatalf("unexpected finding rule: got %q want %q", got, want)
	}
}

func TestLintCommandTargetDirModeSupportsFix(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("ROOT=value  \n\n\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	path := filepath.Join(root, "src", "app", ".env.example")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte("APP=value  \n\n\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--fix", "--target-dir=src")
	if err != nil {
		t.Fatalf("expected lint --fix to succeed, got %v", err)
	}
	if stdout != "" {
		t.Fatalf("expected no output after fixing, got %q", stdout)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != "APP=value\n" {
		t.Fatalf("unexpected fixed content: got %q want %q", string(got), "APP=value\n")
	}

	rootGot, err := os.ReadFile(filepath.Join(root, ".env"))
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(rootGot) != "ROOT=value  \n\n\n" {
		t.Fatalf("unexpected root file content after target-dir fix: %q", string(rootGot))
	}
}

func TestLintCommandTargetDirModeSupportsSkip(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "src", "app", ".env.example")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte("APP=value  \nAPP=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--json", "--skip=duplicate-key", "--target-dir=src")
	if err == nil {
		t.Fatal("expected findings error")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}

	var payload struct {
		Findings []lint.Finding `json:"findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got, want := len(payload.Findings), 1; got != want {
		t.Fatalf("unexpected finding count: got %d want %d", got, want)
	}
	if got, want := payload.Findings[0].File, filepath.Join("src", "app", ".env.example"); got != want {
		t.Fatalf("unexpected finding file: got %q want %q", got, want)
	}
	if got, want := payload.Findings[0].Rule, "trailing-whitespace"; got != want {
		t.Fatalf("unexpected finding rule: got %q want %q", got, want)
	}
}

func TestLintCommandTargetScopeErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("KEY=value\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	cases := []struct {
		name     string
		args     []string
		wantText string
	}{
		{
			name:     "conflicting flags",
			args:     []string{"--target=.env", "--target-dir=src"},
			wantText: "--target and --target-dir cannot be used together",
		},
		{
			name:     "missing target file",
			args:     []string{"--target=missing.env"},
			wantText: "--target path does not exist: missing.env",
		},
		{
			name:     "target points to directory",
			args:     []string{"--target=src"},
			wantText: "--target must point to a file: src",
		},
		{
			name:     "missing target dir",
			args:     []string{"--target-dir=missing-dir"},
			wantText: "--target-dir path does not exist: missing-dir",
		},
		{
			name:     "target dir points to file",
			args:     []string{"--target-dir=.env"},
			wantText: "--target-dir must point to a directory: .env",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runLintCommand(t, root, tc.args...)
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := ExitCode(err); got != ExitInternal {
				t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
			}
			if !strings.Contains(err.Error(), tc.wantText) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestLintCommandTargetFileReportsUnreadablePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-denied fixtures are not portable on windows")
	}

	root := t.TempDir()
	locked := filepath.Join(root, "locked")
	if err := os.Mkdir(locked, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(locked, ".env.staging"), []byte("KEY=value\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(locked, 0o755); err != nil {
			t.Errorf("restore permissions failed: %v", err)
		}
	})
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}

	_, err := runLintCommand(t, root, "--target=locked/.env.staging")
	if err == nil {
		t.Fatal("expected lint --target to fail")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "--target path cannot be read: locked/.env.staging") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLintCommandDoesNotLeakSecretValues(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	secretOne := "supersecret-token-123"
	secretTwo := "another-supersecret-token-456"
	content := "API_TOKEN=" + secretOne + "  \n=" + secretTwo + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	t.Run("text", func(t *testing.T) {
		stdout, err := runLintCommand(t, root)
		if err == nil {
			t.Fatal("expected findings error")
		}
		if got := ExitCode(err); got != ExitFindings {
			t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
		}
		if strings.Contains(stdout, secretOne) || strings.Contains(stdout, secretTwo) {
			t.Fatalf("text output leaked a secret value: %q", stdout)
		}
	})

	t.Run("json", func(t *testing.T) {
		stdout, err := runLintCommand(t, root, "--json")
		if err == nil {
			t.Fatal("expected findings error")
		}
		if got := ExitCode(err); got != ExitFindings {
			t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
		}
		if strings.Contains(stdout, secretOne) || strings.Contains(stdout, secretTwo) {
			t.Fatalf("json output leaked a secret value: %q", stdout)
		}

		var payload struct {
			Findings []lint.Finding `json:"findings"`
		}
		if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if got, want := len(payload.Findings), 2; got != want {
			t.Fatalf("unexpected finding count: got %d want %d", got, want)
		}
	})
}

func TestLintCommandReturnsInternalErrorWhenDiscoveryFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-denied discovery fixtures are not portable on windows")
	}

	root := t.TempDir()
	locked := filepath.Join(root, "locked")
	if err := os.Mkdir(locked, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(locked, ".env"), []byte("KEY=value\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(locked, 0o755); err != nil {
			t.Errorf("restore permissions failed: %v", err)
		}
	})
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}

	_, err := runLintCommand(t, root)
	if err == nil {
		t.Fatal("expected discovery to fail")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Errorf("restore working dir failed: %v", err)
		}
	})
}

func runLintCommand(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()

	withWorkingDir(t, root)

	var stdout bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetArgs(append([]string{"lint"}, args...))

	err := cmd.Execute()
	return stdout.String(), err
}
