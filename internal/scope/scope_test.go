/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package scope

import (
	"os"
	"path/filepath"
	reflect "reflect"
	"runtime"
	"strings"
	"testing"
)

func TestResolveDefaultDiscovery(t *testing.T) {
	root := t.TempDir()
	withWorkingDir(t, root)
	resolvedRoot := mustAbs(t, ".")

	mustWrite(t, filepath.Join(resolvedRoot, ".env"), "ROOT=value\n")
	mustWrite(t, filepath.Join(resolvedRoot, ".env.preview-local"), "PREVIEW=value\n")
	mustWrite(t, filepath.Join(resolvedRoot, "src", "app", ".env.example"), "APP=value\n")
	mustWrite(t, filepath.Join(resolvedRoot, "src", "examples", "broken", ".env.example"), "BROKEN=value\n")
	mustWrite(t, filepath.Join(resolvedRoot, "src", "dist", ".env.local"), "IGNORED=value\n")
	mustWrite(t, filepath.Join(resolvedRoot, "vendor", ".env"), "IGNORED=value\n")

	selection, err := Resolve(Options{Root: "."})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if got, want := selection.Root, resolvedRoot; got != want {
		t.Fatalf("unexpected root: got %q want %q", got, want)
	}
	if got, want := selection.RootLabel, "."; got != want {
		t.Fatalf("unexpected root label: got %q want %q", got, want)
	}

	wantPaths := []string{
		filepath.Join(resolvedRoot, ".env"),
		filepath.Join(resolvedRoot, ".env.preview-local"),
		filepath.Join(resolvedRoot, "src", "app", ".env.example"),
		filepath.Join(resolvedRoot, "src", "examples", "broken", ".env.example"),
	}
	if !reflect.DeepEqual(selection.Paths, wantPaths) {
		t.Fatalf("unexpected paths: got %#v want %#v", selection.Paths, wantPaths)
	}

	wantDisplay := []string{
		".env",
		".env.preview-local",
		filepath.Join("src", "app", ".env.example"),
		filepath.Join("src", "examples", "broken", ".env.example"),
	}
	gotDisplay := make([]string, len(selection.Paths))
	for i, path := range selection.Paths {
		gotDisplay[i] = selection.DisplayPath(path)
	}
	if !reflect.DeepEqual(gotDisplay, wantDisplay) {
		t.Fatalf("unexpected display paths: got %#v want %#v", gotDisplay, wantDisplay)
	}
}

func TestResolveTargetFileSelection(t *testing.T) {
	root := t.TempDir()
	withWorkingDir(t, root)
	resolvedRoot := mustAbs(t, ".")

	mustWrite(t, filepath.Join(resolvedRoot, ".env"), "ROOT=value\n")
	customPath := filepath.Join(resolvedRoot, "config", "custom.envfile")
	mustWrite(t, customPath, "CUSTOM=value\n")

	cases := []struct {
		name        string
		target      string
		wantPath    string
		wantDisplay string
	}{
		{
			name:        "relative target path",
			target:      filepath.Join("config", "custom.envfile"),
			wantPath:    customPath,
			wantDisplay: filepath.Join("config", "custom.envfile"),
		},
		{
			name:        "absolute target path",
			target:      customPath,
			wantPath:    customPath,
			wantDisplay: filepath.Join("config", "custom.envfile"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			selection, err := Resolve(Options{Root: ".", Target: tc.target})
			if err != nil {
				t.Fatalf("resolve failed: %v", err)
			}

			if got, want := len(selection.Paths), 1; got != want {
				t.Fatalf("unexpected path count: got %d want %d", got, want)
			}
			if got, want := selection.Paths[0], tc.wantPath; got != want {
				t.Fatalf("unexpected path: got %q want %q", got, want)
			}
			if got, want := selection.DisplayPath(selection.Paths[0]), tc.wantDisplay; got != want {
				t.Fatalf("unexpected display path: got %q want %q", got, want)
			}
		})
	}
}

func TestResolveTargetDirSelection(t *testing.T) {
	root := t.TempDir()
	withWorkingDir(t, root)
	resolvedRoot := mustAbs(t, ".")

	mustWrite(t, filepath.Join(resolvedRoot, ".env"), "ROOT=value\n")
	mustWrite(t, filepath.Join(resolvedRoot, "src", "app", ".env.example"), "APP=value\n")
	mustWrite(t, filepath.Join(resolvedRoot, "src", "examples", "broken", ".env.example"), "BROKEN=value\n")
	mustWrite(t, filepath.Join(resolvedRoot, "src", "dist", ".env.local"), "IGNORED=value\n")
	mustWrite(t, filepath.Join(resolvedRoot, "src", "README.md"), "ignored\n")
	wantPaths := []string{
		filepath.Join(resolvedRoot, "src", "app", ".env.example"),
		filepath.Join(resolvedRoot, "src", "examples", "broken", ".env.example"),
	}

	cases := []struct {
		name        string
		targetDir   string
		wantDisplay []string
	}{
		{
			name:      "relative target dir",
			targetDir: "src",
			wantDisplay: []string{
				filepath.Join("src", "app", ".env.example"),
				filepath.Join("src", "examples", "broken", ".env.example"),
			},
		},
		{
			name:      "absolute target dir",
			targetDir: filepath.Join(resolvedRoot, "src"),
			wantDisplay: []string{
				filepath.Join("src", "app", ".env.example"),
				filepath.Join("src", "examples", "broken", ".env.example"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			selection, err := Resolve(Options{Root: ".", TargetDir: tc.targetDir})
			if err != nil {
				t.Fatalf("resolve failed: %v", err)
			}

			if !reflect.DeepEqual(selection.Paths, wantPaths) {
				t.Fatalf("unexpected paths: got %#v want %#v", selection.Paths, wantPaths)
			}

			gotDisplay := make([]string, len(selection.Paths))
			for i, path := range selection.Paths {
				gotDisplay[i] = selection.DisplayPath(path)
			}
			if !reflect.DeepEqual(gotDisplay, tc.wantDisplay) {
				t.Fatalf("unexpected display paths: got %#v want %#v", gotDisplay, tc.wantDisplay)
			}
		})
	}
}

func TestResolveErrors(t *testing.T) {
	root := t.TempDir()
	withWorkingDir(t, root)

	mustWrite(t, filepath.Join(root, ".env"), "ROOT=value\n")
	if err := os.Mkdir(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	cases := []struct {
		name    string
		opts    Options
		wantErr string
	}{
		{
			name:    "conflicting target flags",
			opts:    Options{Root: ".", Target: ".env", TargetDir: "src"},
			wantErr: "--target and --target-dir cannot be used together",
		},
		{
			name:    "missing target file",
			opts:    Options{Root: ".", Target: "missing.env"},
			wantErr: "--target path does not exist: missing.env",
		},
		{
			name:    "missing target dir",
			opts:    Options{Root: ".", TargetDir: "missing-dir"},
			wantErr: "--target-dir path does not exist: missing-dir",
		},
		{
			name:    "target points to directory",
			opts:    Options{Root: ".", Target: "src"},
			wantErr: "--target must point to a file: src",
		},
		{
			name:    "target-dir points to file",
			opts:    Options{Root: ".", TargetDir: ".env"},
			wantErr: "--target-dir must point to a directory: .env",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Resolve(tc.opts)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("unexpected error: got %v want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestResolveUnreadableTargetPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-denied fixtures are not portable on windows")
	}
	if isPrivilegedUser() {
		t.Skip("permission-denied fixtures require a non-root POSIX user")
	}

	root := t.TempDir()
	withWorkingDir(t, root)

	locked := filepath.Join(root, "locked")
	if err := os.Mkdir(locked, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	mustWrite(t, filepath.Join(locked, ".env.staging"), "KEY=value\n")
	t.Cleanup(func() {
		if err := os.Chmod(locked, 0o755); err != nil {
			t.Errorf("restore permissions failed: %v", err)
		}
	})
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}

	_, err := Resolve(Options{Root: ".", Target: filepath.Join("locked", ".env.staging")})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "--target path cannot be read: locked/.env.staging") {
		t.Fatalf("unexpected error: %v", err)
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

func mustWrite(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}

func mustAbs(t *testing.T, path string) string {
	t.Helper()

	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs failed: %v", err)
	}
	return abs
}
