/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import "errors"

const (
	// ExitOK means the command completed without findings or errors.
	ExitOK = 0
	// ExitFindings means the command found lint issues and should return a
	// non-zero but expected status.
	ExitFindings = 1
	// ExitInternal means the command hit an unexpected failure.
	ExitInternal = 2
)

// ExitError carries a controlled process exit code and an optional cause.
type ExitError struct {
	// Code is the process exit status to return.
	Code int
	// Err keeps the wrapped cause when one exists.
	Err error
}

func (e ExitError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e ExitError) Unwrap() error {
	return e.Err
}

// IsExitError reports whether err already carries a controlled exit code.
func IsExitError(err error) bool {
	var exitErr ExitError
	return errors.As(err, &exitErr)
}

// ExitCode returns the process exit code implied by err, falling back to the
// internal-error code for unexpected failures.
func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}

	var exitErr ExitError
	if errors.As(err, &exitErr) {
		if exitErr.Code == 0 {
			return ExitOK
		}
		return exitErr.Code
	}

	return ExitInternal
}
