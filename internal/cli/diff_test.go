/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffCommandJSONWithDifferences(t *testing.T) {
	root := t.TempDir()

	left := filepath.Join(root, ".env")
	right := filepath.Join(root, ".env.example")

	if err := os.WriteFile(left, []byte("FOO=secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(right, []byte("BAR=value\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer

	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{
		"diff",
		left,
		right,
		"--json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected difference exit")
	}

	if got := ExitCode(err); got != ExitFindings {
		t.Fatalf("got exit code %d want %d", got, ExitFindings)
	}

	var output diffJSON

	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(output.MissingFromLeft) != 1 ||
		output.MissingFromLeft[0] != "BAR" {
		t.Fatalf("unexpected missing_from_left: %#v", output.MissingFromLeft)
	}

	if len(output.MissingFromRight) != 1 ||
		output.MissingFromRight[0] != "FOO" {
		t.Fatalf("unexpected missing_from_right: %#v", output.MissingFromRight)
	}

	if !output.Different {
		t.Fatal("expected different=true")
	}

	if strings.Contains(stdout.String(), "secret") {
		t.Fatal("JSON leaked value")
	}
}

func TestDiffCommandJSONMatchingFiles(t *testing.T) {
	root := t.TempDir()

	left := filepath.Join(root, ".env")
	right := filepath.Join(root, ".env.example")

	os.WriteFile(left, []byte("FOO=123\n"), 0o644)
	os.WriteFile(right, []byte("FOO=secret\n"), 0o644)

	var stdout bytes.Buffer

	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{
		"diff",
		left,
		right,
		"--json",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output diffJSON

	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatal(err)
	}

	if output.Different {
		t.Fatal("expected different=false")
	}

	if strings.Contains(stdout.String(), "123") ||
		strings.Contains(stdout.String(), "secret") {
		t.Fatal("JSON leaked values")
	}
}

func TestDiffCommandRejectsQuietWithJSON(t *testing.T) {
	cmd := newRootCmd()

	cmd.SetArgs([]string{
		"diff",
		"a.env",
		"b.env",
		"--json",
		"--quiet",
	})

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "--quiet and --json cannot be used together") {
		t.Fatalf("unexpected error: %v", err)
	}
}
