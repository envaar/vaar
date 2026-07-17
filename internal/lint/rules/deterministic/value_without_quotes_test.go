/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestValueWithoutQuotes(t *testing.T) {
	tests := []ruleTestCase{
		{
			rule:            deterministic.NewValueWithoutQuotes(),
			input:           []byte("FOO=BAR BAZ\n"),
			wantCount:       1,
			wantLine:        1,
			wantRule:        "value-without-quotes",
			wantSeverity:    "error",
			wantMessagePart: "value containing whitespace should be enclosed in quotes",
		},
		{
			rule:      deterministic.NewValueWithoutQuotes(),
			input:     []byte("WELCOME_MESSAGE='Hello world'\n"),
			wantCount: 0,
		},
		{
			rule:      deterministic.NewValueWithoutQuotes(),
			input:     []byte("DATABASE_LABEL=\"Primary database\"\n"),
			wantCount: 0,
		},
		{
			rule:      deterministic.NewValueWithoutQuotes(),
			input:     []byte("FOO=BAR\n"),
			wantCount: 0,
		},
		{
			rule:      deterministic.NewValueWithoutQuotes(),
			input:     []byte("EMPTY=\n"),
			wantCount: 0,
		},
		{
			rule:            deterministic.NewValueWithoutQuotes(),
			input:           []byte("MULTIPLE_SPACES=hello   world\n"),
			wantCount:       1,
			wantLine:        1,
			wantRule:        "value-without-quotes",
			wantSeverity:    "error",
			wantMessagePart: "value containing whitespace should be enclosed in quotes",
		},
		{
			rule:            deterministic.NewValueWithoutQuotes(),
			input:           []byte("TAB_VALUE=hello\tworld\n"),
			wantCount:       1,
			wantLine:        1,
			wantRule:        "value-without-quotes",
			wantSeverity:    "error",
			wantMessagePart: "value containing whitespace should be enclosed in quotes",
		},
		{
			rule:      deterministic.NewValueWithoutQuotes(),
			input:     []byte("INLINE_COMMENT=value # comment\n"),
			wantCount: 0,
		},
		{
			rule:      deterministic.NewValueWithoutQuotes(),
			input:     []byte("HASH_IN_QUOTES=\"value # not a comment\"\n"),
			wantCount: 0,
		},
		{
			rule:      deterministic.NewValueWithoutQuotes(),
			input:     []byte("BROKEN=\"hello world\n"),
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		runRuleTest(t, tc)
	}
}
