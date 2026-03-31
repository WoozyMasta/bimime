// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

/*
Package bimime provides a unified registry of BI game file types with MIME
names, descriptions, binary/text hints, extension/path matching, and
magic-byte probing.

Use Analyze/AnalyzeReader/AnalyzeFile with AnalyzeOptions and AnalyzePlan:
  - extension-only matching for fast path.
  - extension+magic when content is needed.
  - strict validation when consistency checks are required.
  - extension-specific plan overrides for mixed corpora.
  - fast + forced magic for ambiguous RAP-like extensions
    via BIAmbiguousRAPOptions.

Magic-byte detection is considered more reliable than extension-only detection.
Use Probe when both filename and payload prefix are available.
*/
package bimime
