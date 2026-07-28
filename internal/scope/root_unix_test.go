//go:build !windows

/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package scope

import "os"

func isPrivilegedUser() bool {
	return os.Geteuid() == 0
}
