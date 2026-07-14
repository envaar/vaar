/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint

import (
	"bytes"
	"os"

	"github.com/envaar/vaar/internal/envfile"
)

// ApplyFixes normalizes each discovered dotenv file and reports whether any
// file changed on disk.
func ApplyFixes(paths []string) (bool, error) {
	changed := false
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}

		fixed := envfile.Normalize(data)
		if bytes.Equal(fixed, data) {
			continue
		}

		if err := envfile.Write(path, fixed); err != nil {
			return false, err
		}
		changed = true
	}

	return changed, nil
}
