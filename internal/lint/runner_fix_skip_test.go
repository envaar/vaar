package lint

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type runCountRule struct {
	calls *int
}

func (r runCountRule) ID() string { return "run-count" }

func (r runCountRule) Run(Context) ([]Finding, error) {
	*r.calls++
	return nil, nil
}

func TestRunnerFixSkipsSecondLintPassWhenNoFilesChanged(t *testing.T) {
	root := t.TempDir()
	envPath := filepath.Join(root, ".env")
	if err := os.WriteFile(envPath, []byte("KEY=value\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	calls := 0
	runner := NewRunner(runCountRule{calls: &calls})

	_, err := runner.Run(context.Background(), Options{
		Root: root,
		Fix:  true,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if calls != 1 {
		t.Fatalf("expected one lint pass when --fix makes no changes, got %d", calls)
	}
}
