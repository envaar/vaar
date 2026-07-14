/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package cli verifies the root command's version output and Cobra wiring.
package cli

import (
	"bytes"
	"testing"
)

func TestRootCommandVersionFlag(t *testing.T) {
	var stdout bytes.Buffer

	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if got, want := stdout.String(), "vaar version dev\n"; got != want {
		t.Fatalf("unexpected version output: got %q want %q", got, want)
	}
}
