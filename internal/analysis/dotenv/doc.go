// Copyright © 2026 envaar
// SPDX-License-Identifier: Apache-2.0

// Package dotenv converts loaded dotenv source documents into the core
// value-free analysis model. It preserves document and line order, identity,
// provenance, and approved parser facts without performing filesystem I/O,
// parsing, validation, indexing, engine execution, output, or mutation.
//
// Parser values and other raw source data are read only transiently when
// deriving safe facts and are never retained in the analysis document.
package dotenv
