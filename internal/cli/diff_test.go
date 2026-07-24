/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestDiffCommandReportsDifferencesForRelativePaths(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\nCOMMON=local\n")
	mustWriteFile(t, filepath.Join(root, ".env.example"), "BAR=example\nCOMMON=example\n")

	stdout, stderr, err := runDiffCommandWithStreams(t, root, ".env", ".env.example")
	if err == nil {
		t.Fatal("expected differences")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}
	want := ".env is missing key: BAR\n.env.example is missing key: FOO\n"
	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
}

func TestDiffCommandReportsDifferencesForAbsolutePaths(t *testing.T) {
	root := t.TempDir()
	left := filepath.Join(root, "left.env")
	right := filepath.Join(root, "right.env")
	mustWriteFile(t, left, "DEBUG=true\nFOO=local\n")
	mustWriteFile(t, right, "DATABASE_URL=postgres://example\nBAR=example\nBAZ=example\n")

	stdout, stderr, err := runDiffCommandWithStreams(t, root, left, right)
	if err == nil {
		t.Fatal("expected differences")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}
	want := left + " is missing keys: BAR, BAZ, DATABASE_URL\n" +
		right + " is missing keys: DEBUG, FOO\n"
	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
}

func TestDiffCommandSupportsPathsContainingSpaces(t *testing.T) {
	root := t.TempDir()
	left := filepath.Join(root, "local env")
	right := filepath.Join(root, "example env")
	mustWriteFile(t, left, "FOO=local\n")
	mustWriteFile(t, right, "BAR=example\n")

	stdout, stderr, err := runDiffCommandWithStreams(t, root, "local env", "example env")
	if err == nil {
		t.Fatal("expected differences")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}
	want := "local env is missing key: BAR\nexample env is missing key: FOO\n"
	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
}

func TestDiffCommandMatchingFilesExitZero(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\nCOMMON=local\n")
	mustWriteFile(t, filepath.Join(root, ".env.example"), "COMMON=example\nFOO=example\n")

	stdout, stderr, err := runDiffCommandWithStreams(t, root, ".env", ".env.example")
	if err != nil {
		t.Fatalf("expected matching files to succeed, got %v", err)
	}
	if got := ExitCode(err); got != ExitOK {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitOK)
	}
	if want := "No key differences found\n"; stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
}

func TestDiffCommandQuietSuppressesDifferenceOutput(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\nCOMMON=local\n")
	mustWriteFile(t, filepath.Join(root, ".env.example"), "BAR=example\nCOMMON=example\n")

	stdout, stderr, err := runDiffCommandWithStreams(t, root, ".env", ".env.example", "--quiet")
	if err == nil {
		t.Fatal("expected differences")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
}

func TestDiffCommandQuietSuppressesSuccessOutput(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\n")
	mustWriteFile(t, filepath.Join(root, ".env.example"), "FOO=example\n")

	stdout, stderr, err := runDiffCommandWithStreams(t, root, ".env", ".env.example", "--quiet")
	if err != nil {
		t.Fatalf("expected matching files to succeed, got %v", err)
	}
	if got := ExitCode(err); got != ExitOK {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitOK)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
}

func TestDiffCommandQuietStillReportsOperationalFailures(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\n")

	stdout, stderr, err := runDiffCommandWithStreams(t, root, ".env", ".env.example", "--quiet")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout, got %q", stdout)
	}
	want := "error: reading .env.example: file does not exist\n"
	if stderr != want {
		t.Fatalf("unexpected stderr: got %q want %q", stderr, want)
	}
}

func TestDiffCommandJSONWithDifferences(t *testing.T) {
	root := t.TempDir()
	left := filepath.Join(root, ".env")
	right := filepath.Join(root, ".env.example")
	mustWriteFile(t, left, "ZOO=left-secret\nFOO=left-token\nCOMMON=left\n")
	mustWriteFile(t, right, "BAR=right-value\nAAA=right-secret\nCOMMON=right\n")

	stdout, stderr, err := runDiffCommandWithStreams(t, root, left, right, "--json")
	if err == nil {
		t.Fatal("expected differences")
	}
	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitFindings)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}

	output := parseDiffJSON(t, stdout)
	if output.Left != left || output.Right != right {
		t.Fatalf("unexpected paths: got left=%q right=%q", output.Left, output.Right)
	}
	if !slices.Equal(output.MissingFromLeft, []string{"AAA", "BAR"}) {
		t.Fatalf("unexpected missing_from_left: %#v", output.MissingFromLeft)
	}
	if !slices.Equal(output.MissingFromRight, []string{"FOO", "ZOO"}) {
		t.Fatalf("unexpected missing_from_right: %#v", output.MissingFromRight)
	}
	if !output.Different {
		t.Fatal("expected different=true")
	}
	for _, value := range []string{"left-secret", "left-token", "right-value", "right-secret"} {
		if strings.Contains(stdout, value) {
			t.Fatalf("JSON leaked value %q in %q", value, stdout)
		}
	}
}

func TestDiffCommandJSONMatchingFiles(t *testing.T) {
	root := t.TempDir()
	left := filepath.Join(root, ".env")
	right := filepath.Join(root, ".env.example")
	mustWriteFile(t, left, "FOO=123\nCOMMON=local-secret\n")
	mustWriteFile(t, right, "COMMON=example-secret\nFOO=456\n")

	stdout, stderr, err := runDiffCommandWithStreams(t, root, left, right, "--json")
	if err != nil {
		t.Fatalf("expected matching files to succeed, got %v", err)
	}
	if got := ExitCode(err); got != ExitOK {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitOK)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}

	output := parseDiffJSON(t, stdout)
	if output.Left != left || output.Right != right {
		t.Fatalf("unexpected paths: got left=%q right=%q", output.Left, output.Right)
	}
	if len(output.MissingFromLeft) != 0 {
		t.Fatalf("unexpected missing_from_left: %#v", output.MissingFromLeft)
	}
	if len(output.MissingFromRight) != 0 {
		t.Fatalf("unexpected missing_from_right: %#v", output.MissingFromRight)
	}
	if output.Different {
		t.Fatal("expected different=false")
	}
	for _, value := range []string{"123", "456", "local-secret", "example-secret"} {
		if strings.Contains(stdout, value) {
			t.Fatalf("JSON leaked value %q in %q", value, stdout)
		}
	}
}

func TestDiffCommandRejectsQuietWithJSON(t *testing.T) {
	root := t.TempDir()

	stdout, stderr, err := runDiffCommandWithStreams(t, root, "a.env", "b.env", "--json", "--quiet")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if got, want := err.Error(), "--quiet and --json cannot be used together"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout, got %q", stdout)
	}
	if want := "error: --quiet and --json cannot be used together\n"; stderr != want {
		t.Fatalf("unexpected stderr: got %q want %q", stderr, want)
	}
}

func TestDiffCommandRejectsWrongArgumentCount(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\n")
	mustWriteFile(t, filepath.Join(root, ".env.example"), "FOO=example\n")

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "too few",
			args: []string{".env"},
		},
		{
			name: "too many",
			args: []string{".env", ".env.example", ".env.production"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runDiffCommandWithStreams(t, root, tc.args...)
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := ExitCode(err); got != ExitInternal {
				t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
			}
			if got, want := err.Error(), "diff requires exactly two dotenv files"; got != want {
				t.Fatalf("unexpected error: got %q want %q", got, want)
			}
			if stdout != "" {
				t.Fatalf("expected no stdout, got %q", stdout)
			}
			if want := "error: diff requires exactly two dotenv files\n"; stderr != want {
				t.Fatalf("unexpected stderr: got %q want %q", stderr, want)
			}
		})
	}
}

func TestDiffCommandReportsOperationalFailures(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\n")
	if err := os.Mkdir(filepath.Join(root, "configs"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		wantText string
	}{
		{
			name:     "missing input",
			args:     []string{".env", ".env.example"},
			wantText: "reading .env.example: file does not exist",
		},
		{
			name:     "directory input",
			args:     []string{".env", "configs"},
			wantText: "configs is a directory, expected a dotenv file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runDiffCommandWithStreams(t, root, tc.args...)
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := ExitCode(err); got != ExitInternal {
				t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
			}
			if !strings.Contains(err.Error(), tc.wantText) {
				t.Fatalf("unexpected error: %v", err)
			}
			if stdout != "" {
				t.Fatalf("expected no stdout, got %q", stdout)
			}
			wantStderr := "error: " + tc.wantText + "\n"
			if stderr != wantStderr {
				t.Fatalf("unexpected stderr: got %q want %q", stderr, wantStderr)
			}
		})
	}
}

func TestDiffCommandRejectsNonRegularInput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("non-regular /dev/null fixture is Unix-specific")
	}

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\n")

	stdout, stderr, err := runDiffCommandWithStreams(t, root, ".env", "/dev/null")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	wantText := "/dev/null is not a regular file, expected a dotenv file"
	if !strings.Contains(err.Error(), wantText) {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout, got %q", stdout)
	}
	if stderr != "error: "+wantText+"\n" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestDiffCommandReportsUnreadableFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-denied fixtures are not portable on windows")
	}

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\n")
	locked := filepath.Join(root, ".env.example")
	mustWriteFile(t, locked, "FOO=example\n")
	t.Cleanup(func() {
		if err := os.Chmod(locked, 0o644); err != nil {
			t.Errorf("restore permissions failed: %v", err)
		}
	})
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}
	if file, err := os.Open(locked); err == nil {
		_ = file.Close()
		t.Skip("permission-denied fixture remains readable, likely running with elevated privileges")
	}

	stdout, stderr, err := runDiffCommandWithStreams(t, root, ".env", ".env.example")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "reading .env.example") {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Fatalf("expected no stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "error: reading .env.example") {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestDiffCommandPropagatesStdoutWriteFailures(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\nCOMMON=local\n")
	mustWriteFile(t, filepath.Join(root, ".env.example"), "BAR=example\nCOMMON=example\n")

	withWorkingDir(t, root)

	cmd := newRootCmd()
	cmd.SetOut(failingWriter{})
	cmd.SetArgs([]string{"diff", ".env", ".env.example"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "writing diff output failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDiffCommandPropagatesSuccessStdoutWriteFailures(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\n")
	mustWriteFile(t, filepath.Join(root, ".env.example"), "FOO=example\n")

	withWorkingDir(t, root)

	cmd := newRootCmd()
	cmd.SetOut(failingWriter{})
	cmd.SetArgs([]string{"diff", ".env", ".env.example"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "writing diff output failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDiffCommandPropagatesJSONStdoutWriteFailures(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, ".env"), "FOO=local\n")
	mustWriteFile(t, filepath.Join(root, ".env.example"), "FOO=example\n")

	withWorkingDir(t, root)

	cmd := newRootCmd()
	cmd.SetOut(failingWriter{})
	cmd.SetArgs([]string{"diff", ".env", ".env.example", "--json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := ExitCode(err); got != ExitInternal {
		t.Fatalf("unexpected exit code: got %d want %d", got, ExitInternal)
	}
	if !strings.Contains(err.Error(), "writing diff output failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func runDiffCommandWithStreams(t *testing.T, root string, args ...string) (string, string, error) {
	t.Helper()

	withWorkingDir(t, root)

	var stdout, stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(append([]string{"diff"}, args...))

	err := cmd.Execute()
	if err != nil && !IsExitError(err) {
		writeTestCLIError(&stderr, err)
	}
	return stdout.String(), stderr.String(), err
}

func parseDiffJSON(t *testing.T, output string) diffJSON {
	t.Helper()

	var payload diffJSON
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return payload
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s failed: %v", path, err)
	}
}

func writeTestCLIError(w *bytes.Buffer, err error) {
	_, _ = w.WriteString("error: " + err.Error() + "\n")
}

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}
