// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

/*
Package bimime provides a unified registry of BI game file types with MIME
names, descriptions, binary/text hints, extension/path matching, and
magic-byte probing.

Use Analyze/AnalyzeReader/AnalyzeFile with mode:
  - fast: extension/path-hint only, no payload reads.
  - normal: extension/path-hint and magic when content is needed.
  - strict: normal + consistency checks (magic and quick text validation).

Magic-byte detection is considered more reliable than extension-only detection.
Use Probe when both filename and payload prefix are available.
*/
package bimime
