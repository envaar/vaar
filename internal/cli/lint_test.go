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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules"
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

	stdout, stderr, err := runLintCommandWithStreams(t, root, "--json", "--output=lint-report.json")
	if err == nil {
		t.Fatal("expected findings error")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout output, got %q", stdout)
	}
	wantStderr := "Successfully flagged 1 finding and wrote to lint-report.json\n"
	if stderr != wantStderr {
		t.Fatalf("unexpected stderr: got %q want %q", stderr, wantStderr)
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
	wantOutput := "[fixed] warn trailing-whitespace .env:1 line has trailing whitespace\n" +
		"[fixed] warn ending-blank-line .env:3 file must end with exactly one final newline\n" +
		"[fixed] warn extra-blank-line .env:3 repeated blank line\n"
	if stdout != wantOutput {
		t.Fatalf("unexpected fix report: got %q want %q", stdout, wantOutput)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != "KEY=value\n" {
		t.Fatalf("unexpected fixed content: got %q want %q", string(got), "KEY=value\n")
	}
}

func TestLintCommandFixUsesRemainingFindingsForExitStatus(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("KEY=value  \nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--fix")
	if err == nil {
		t.Fatal("expected the remaining duplicate-key finding to fail the command")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}
	wantOutput := "[fixed] warn trailing-whitespace .env:1 line has trailing whitespace\n" +
		"error duplicate-key .env:2 KEY is defined more than once\n"
	if stdout != wantOutput {
		t.Fatalf("unexpected fix report: got %q want %q", stdout, wantOutput)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != "KEY=value\nKEY=other\n" {
		t.Fatalf("unexpected fixed content: got %q", string(got))
	}
}

func TestLintCommandFixKeepsShiftedFindingsVisible(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("KEY=value  \n\n\nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--fix")
	if err == nil {
		t.Fatal("expected the remaining duplicate-key finding to fail the command")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}
	wantOutput := "[fixed] warn trailing-whitespace .env:1 line has trailing whitespace\n" +
		"error duplicate-key .env:3 KEY is defined more than once\n" +
		"[fixed] warn extra-blank-line .env:3 repeated blank line\n"
	if stdout != wantOutput {
		t.Fatalf("unexpected fix report: got %q want %q", stdout, wantOutput)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != "KEY=value\n\nKEY=other\n" {
		t.Fatalf("unexpected fixed content: got %q", string(got))
	}
}

func TestLintCommandFixMarksFindingsInJSON(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("KEY=value  \nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--fix", "--json")
	if err == nil {
		t.Fatal("expected the remaining duplicate-key finding to fail the command")
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
	if !payload.Findings[0].Fixed {
		t.Fatalf("expected first finding to be marked fixed: %#v", payload.Findings[0])
	}
	if payload.Findings[1].Fixed {
		t.Fatalf("expected remaining finding not to be marked fixed: %#v", payload.Findings[1])
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
	expectedPath := filepath.Join("src", "app", ".env.example")
	wantOutput := "[fixed] warn trailing-whitespace " + expectedPath + ":1 line has trailing whitespace\n" +
		"[fixed] warn ending-blank-line " + expectedPath + ":3 file must end with exactly one final newline\n" +
		"[fixed] warn extra-blank-line " + expectedPath + ":3 repeated blank line\n"
	if stdout != wantOutput {
		t.Fatalf("unexpected fix report: got %q want %q", stdout, wantOutput)
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

	stdout, _, err := runLintCommandWithStreams(t, root, args...)
	return stdout, err
}

func runLintCommandWithStreams(t *testing.T, root string, args ...string) (string, string, error) {
	t.Helper()

	withWorkingDir(t, root)

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(append([]string{"lint"}, args...))

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestLintCommandRejectsOutputOverlappingInput(t *testing.T) {
	root := t.TempDir()
	envPath := filepath.Join(root, ".env")
	if err := os.WriteFile(envPath, []byte("KEY=value\nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	srcAppPath := filepath.Join(root, "src", "app", ".env")
	if err := os.MkdirAll(filepath.Dir(srcAppPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(srcAppPath, []byte("APP=value\nAPP=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	cases := []struct {
		name     string
		args     []string
		wantPath string
	}{
		{
			name:     "output equals explicit target",
			args:     []string{"--target=.env", "--json", "--output=.env"},
			wantPath: ".env",
		},
		{
			name:     "output equals discovered file",
			args:     []string{"--json", "--output=.env"},
			wantPath: ".env",
		},
		{
			name:     "output equals file under target-dir",
			args:     []string{"--target-dir=src", "--json", "--output=src/app/.env"},
			wantPath: "src/app/.env",
		},
		{
			name:     "relative output matches absolute target",
			args:     []string{"--target=" + envPath, "--json", "--output=.env"},
			wantPath: ".env",
		},
		{
			name:     "output via parent traversal matches discovered file",
			args:     []string{"--json", "--output=src/../.env"},
			wantPath: "src/../.env",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			original, err := os.ReadFile(envPath)
			if err != nil {
				t.Fatalf("read failed: %v", err)
			}

			stdout, err := runLintCommand(t, root, tc.args...)
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := ExitCode(err); got != ExitInternal {
				t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("cannot write lint output to %q", tc.wantPath)) {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(err.Error(), "the path is also a lint input file") {
				t.Fatalf("unexpected error: %v", err)
			}
			if stdout != "" {
				t.Fatalf("expected no stdout output, got %q", stdout)
			}

			after, err := os.ReadFile(envPath)
			if err != nil {
				t.Fatalf("read failed: %v", err)
			}
			if !bytes.Equal(after, original) {
				t.Fatalf("input file was modified: got %q want %q", after, original)
			}
		})
	}
}

func TestLintCommandWritesJSONOutputToDistinctPath(t *testing.T) {
	root := t.TempDir()
	envPath := filepath.Join(root, ".env")
	if err := os.WriteFile(envPath, []byte("KEY=value\nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, stderr, err := runLintCommandWithStreams(t, root, "--json", "--output=lint-report.json")
	if err == nil {
		t.Fatal("expected findings error")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout output, got %q", stdout)
	}
	wantStderr := "Successfully flagged 1 finding and wrote to lint-report.json\n"
	if stderr != wantStderr {
		t.Fatalf("unexpected stderr: got %q want %q", stderr, wantStderr)
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

	after, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(after) != "KEY=value\nKEY=other\n" {
		t.Fatalf("input file was modified: %q", after)
	}
}

func TestLintCommandRejectsOutputThatIsExistingDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("KEY=value\nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	victim := filepath.Join(root, "victim")
	if err := os.Mkdir(victim, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	stdout, err := runLintCommand(t, root, "--json", "--output=victim")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "the path is a directory") {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout output, got %q", stdout)
	}

	// The directory must still exist as a directory and no output must be written.
	info, statErr := os.Stat(victim)
	if statErr != nil {
		t.Fatalf("victim directory was destroyed: %v", statErr)
	}
	if !info.IsDir() {
		t.Fatalf("victim is no longer a directory")
	}
	entries, err := os.ReadDir(victim)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected victim/ to remain empty, got %d entries", len(entries))
	}
}

func TestLintCommandRejectsOutputWithTrailingSeparator(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("KEY=value\nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := runLintCommand(t, root, "--json", "--output=reports/")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "the path is a directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLintCommandListsRulesAlphabetically(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"lint", "--list-rules"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected --list-rules to succeed, got %v", err)
	}

	output := stdout.String()
	if !strings.HasPrefix(output, "NAME") {
		t.Fatalf("expected output to start with NAME header, got %q", output)
	}
	header, _, _ := strings.Cut(output, "\n")
	if !strings.Contains(header, "DESCRIPTION") {
		t.Fatalf("expected header to contain DESCRIPTION column, got %q", header)
	}

	expected := rules.All()
	wantIDs := make([]string, len(expected))
	for i, rule := range expected {
		wantIDs[i] = rule.ID()
	}
	sort.Strings(wantIDs)

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != len(expected)+1 {
		t.Fatalf("expected %d lines (header + %d rules), got %d", len(expected)+1, len(expected), len(lines))
	}

	descriptions := make(map[string]string, len(expected))
	for _, rule := range expected {
		description := strings.TrimSpace(rule.Description())
		if description == "" {
			t.Fatalf("rule %q has an empty description", rule.ID())
		}
		descriptions[rule.ID()] = description
	}

	gotIDs := make([]string, 0, len(expected))
	for _, line := range lines[1:] {
		id, description, found := strings.Cut(strings.TrimSpace(line), " ")
		if !found {
			t.Fatalf("expected rule line to carry a description: %q", line)
		}
		gotIDs = append(gotIDs, id)

		want, known := descriptions[id]
		if !known {
			t.Fatalf("unexpected rule %q in output", id)
		}
		if got := strings.TrimSpace(description); got != want {
			t.Fatalf("wrong description on the %q line:\ngot  %q\nwant %q", id, got, want)
		}
	}

	if !slicesEqual(gotIDs, wantIDs) {
		t.Fatalf("unexpected rule order or content:\ngot  %v\nwant %v", gotIDs, wantIDs)
	}
}

func TestLintCommandListRulesWorksOutsideRepository(t *testing.T) {
	root := t.TempDir()
	withWorkingDir(t, root)

	stdout, err := runLintCommand(t, root, "--list-rules")
	if err != nil {
		t.Fatalf("expected --list-rules to succeed outside a repository, got %v", err)
	}
	if !strings.Contains(stdout, "duplicate-key") {
		t.Fatalf("expected rule list to contain duplicate-key, got %q", stdout)
	}
}

func TestLintCommandListRulesRejectsExecutionFlags(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "only",
			args: []string{"--list-rules", "--only=duplicate-key"},
			want: "--list-rules cannot be combined with --only",
		},
		{
			name: "skip",
			args: []string{"--list-rules", "--skip=trailing-whitespace"},
			want: "--list-rules cannot be combined with --skip",
		},
		{
			name: "fix",
			args: []string{"--list-rules", "--fix"},
			want: "--list-rules cannot be combined with --fix",
		},
		{
			name: "json",
			args: []string{"--list-rules", "--json"},
			want: "--list-rules cannot be combined with --json",
		},
		{
			name: "output",
			args: []string{"--list-rules", "--output=rules.json"},
			want: "--list-rules cannot be combined with --output",
		},
		{
			name: "target",
			args: []string{"--list-rules", "--target=.env"},
			want: "--list-rules cannot be combined with --target",
		},
		{
			name: "target-dir",
			args: []string{"--list-rules", "--target-dir=src"},
			want: "--list-rules cannot be combined with --target-dir",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			stdout, err := runLintCommand(t, root, tc.args...)
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := ExitCode(err); got != ExitInternal {
				t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unexpected error: got %v, want to contain %q", err, tc.want)
			}
			if stdout != "" {
				t.Fatalf("expected no stdout output, got %q", stdout)
			}
		})
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLintCommandConfirmsOutputWriteWithNoFindings(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("KEY=value\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, stderr, err := runLintCommandWithStreams(t, root, "--json", "--output=lint-report.json")
	if err != nil {
		t.Fatalf("expected lint to succeed, got %v", err)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout output, got %q", stdout)
	}
	wantStderr := "Successfully flagged no findings and wrote to lint-report.json\n"
	if stderr != wantStderr {
		t.Fatalf("unexpected stderr: got %q want %q", stderr, wantStderr)
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
	if got, want := len(payload.Findings), 0; got != want {
		t.Fatalf("unexpected finding count: got %d want %d", got, want)
	}
}

func TestLintCommandWritesJSONOutputToSubdirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("KEY=value\nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "reports"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	_, err := runLintCommand(t, root, "--json", "--output=reports/lint.json")
	if err == nil {
		t.Fatal("expected findings error")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}

	data, err := os.ReadFile(filepath.Join(root, "reports", "lint.json"))
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

	// The temporary file must be co-located with the destination and cleaned up,
	// leaving only the report in the subdirectory.
	entries, err := os.ReadDir(filepath.Join(root, "reports"))
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected only the output file in reports/, got %d entries", len(entries))
	}
}

func TestLintCommandReplacesExistingOutputFileWithoutDestroyingIt(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("KEY=value\nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	outPath := filepath.Join(root, "lint.json")
	if err := os.WriteFile(outPath, []byte("STALE-REPORT"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := runLintCommand(t, root, "--json", "--output=lint.json")
	if err == nil {
		t.Fatal("expected findings error")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("existing output file was destroyed: %v", err)
	}
	if strings.Contains(string(data), "STALE-REPORT") {
		t.Fatalf("expected stale output to be replaced, got %q", string(data))
	}
	var payload struct {
		Findings []lint.Finding `json:"findings"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(payload.Findings) != 1 {
		t.Fatalf("unexpected finding count: %d", len(payload.Findings))
	}
}

func TestLintCommandReportsOutputWriteFailure(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("KEY=value\nKEY=other\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	stdout, stderr, err := runLintCommandWithStreams(t, root, "--json", "--output=missing/dir/lint-report.json")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout output, got %q", stdout)
	}
	if strings.Contains(stderr, "Successfully flagged") {
		t.Fatalf("expected no success message on write failure, got %q", stderr)
	}

	if _, statErr := os.Stat(filepath.Join(root, "missing", "dir", "lint-report.json")); !os.IsNotExist(statErr) {
		t.Fatalf("expected output file not to be created, got stat error %v", statErr)
	}
}
