// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

/*
Package bimime provides a built-in registry of BI game file types
and bounded path/content classification.

Classification uses three evidence-cost profiles.

  - [ProfileFast] uses path evidence only,
  - [ProfileNormal] probes content only for useful refinements,
  - [ProfileThorough] performs bounded content discovery and verification.

Callers that own file or container storage can use RequiredPrefix and [Analyze] without I/O.
[AnalyzeReader] and [AnalyzeFile] are convenience wrappers over the same classification rules.
Results retain path evidence, content evidence, their relation, and any incomplete-probe flags;
classification does not claim that a complete file was semantically validated.
*/
package bimime
