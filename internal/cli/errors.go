/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import "fmt"

// ToolError wraps an error with a message that should go to stderr while
// preserving the original cause for errors.Is and errors.As.
type ToolError struct {
	// Message explains the failure in user-facing language.
	Message string
	// Err keeps the wrapped cause for inspection and unwrapping.
	Err error
}

func (e ToolError) Error() string {
	switch {
	case e.Message != "" && e.Err != nil:
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	case e.Message != "":
		return e.Message
	case e.Err != nil:
		return e.Err.Error()
	default:
		return ""
	}
}

func (e ToolError) Unwrap() error {
	return e.Err
}

// NewToolError wraps a lower-level error in a user-facing message.
func NewToolError(message string, err error) error {
	return ToolError{Message: message, Err: err}
}
