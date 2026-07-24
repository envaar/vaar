/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package envfile

import "os"

// Write writes data back to path and preserves the file's existing permissions
// when the file already exists.
func Write(path string, data []byte) error {
	perm := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	}

	return os.WriteFile(path, data, perm)
}
