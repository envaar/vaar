/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"os"

	"github.com/envaar/vaar/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
