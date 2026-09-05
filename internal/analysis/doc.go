// Copyright © 2026 envaar
// SPDX-License-Identifier: Apache-2.0

// Package analysis owns value-free, ordered document and line facts with ordered
// provenance.
// Engines consume these facts, but analysis does not decide lint severity,
// diff output, query status, or mutation behavior. internal/analysis performs
// no file reads/I/O, parsing, discovery, rule execution, diff comparison,
// query evaluation, file writes, output rendering, serialization/JSON,
// exit-code mapping, source conversion, or caller integration. Source
// conversion lives outside this package. This package must not depend on or
// integrate with internal/envfile, internal/source/dotenv, internal/fs,
// internal/scope, internal/lint, internal/diff, internal/output, internal/cli,
// or Cobra. Raw values, raw lines, raw bytes,
// lengths, hashes, masks, and secret-bearing errors are prohibited.
package analysis
