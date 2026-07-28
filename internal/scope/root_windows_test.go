//go:build windows

/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package scope

func isPrivilegedUser() bool {
	return false
}
