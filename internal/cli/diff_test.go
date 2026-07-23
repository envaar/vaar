/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDiffCommandReportsDifferencesForRelativePaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("FOO=local\nCOMMON=local\n"), 0o644); err != nil {
		t.Fatalf("write left failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env.example"), []byte("BAR=example\nCOMMON=example\n"), 0o644); err != nil {
		t.Fatalf("write right failed: %v", err)
	}

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
	if err := os.WriteFile(left, []byte("DEBUG=true\nFOO=local\n"), 0o644); err != nil {
		t.Fatalf("write left failed: %v", err)
	}
	if err := os.WriteFile(right, []byte("DATABASE_URL=postgres://example\nBAR=example\nBAZ=example\n"), 0o644); err != nil {
		t.Fatalf("write right failed: %v", err)
	}

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
	if err := os.WriteFile(left, []byte("FOO=local\n"), 0o644); err != nil {
		t.Fatalf("write left failed: %v", err)
	}
	if err := os.WriteFile(right, []byte("BAR=example\n"), 0o644); err != nil {
		t.Fatalf("write right failed: %v", err)
	}

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
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("FOO=local\nCOMMON=local\n"), 0o644); err != nil {
		t.Fatalf("write left failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env.example"), []byte("COMMON=example\nFOO=example\n"), 0o644); err != nil {
		t.Fatalf("write right failed: %v", err)
	}

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

func TestDiffCommandRejectsWrongArgumentCount(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("FOO=local\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env.example"), []byte("FOO=example\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

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
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("FOO=local\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
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
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("FOO=local\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

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
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("FOO=local\n"), 0o644); err != nil {
		t.Fatalf("write left failed: %v", err)
	}
	locked := filepath.Join(root, ".env.example")
	if err := os.WriteFile(locked, []byte("FOO=example\n"), 0o644); err != nil {
		t.Fatalf("write right failed: %v", err)
	}
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
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("FOO=local\nCOMMON=local\n"), 0o644); err != nil {
		t.Fatalf("write left failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env.example"), []byte("BAR=example\nCOMMON=example\n"), 0o644); err != nil {
		t.Fatalf("write right failed: %v", err)
	}

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
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("FOO=local\n"), 0o644); err != nil {
		t.Fatalf("write left failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env.example"), []byte("FOO=example\n"), 0o644); err != nil {
		t.Fatalf("write right failed: %v", err)
	}

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
		writeCLIError(&stderr, err)
	}
	return stdout.String(), stderr.String(), err
}

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}
