# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

<!--
## Unreleased

### Added
### Changed
### Removed
-->

## Unreleased

### Changed

* Breaking: replace plan and validation options with profile-based
  classification results that preserve path and content evidence.
* Registry types now declare text/binary traits and compatible refinement
  families explicitly; contextual and pattern refinements are table-driven.
* Structured probes declare their targets
  and byte requirements in one capability catalog;
  Normal refinements run only declared targets.
* Weak PAA/BMP signatures and raw MP3 frame headers verify matching paths
  without enabling global content discovery.
* Thorough analysis reports missing expected content
  as soon as its path verifier has enough bytes,
  independent of the global prefix budget.
* Optional path-pattern refinements no longer imply required content.
* Runtime classification uses compiled registry metadata;
  contextual targets must match or refine their path type family.
* `ByContent` preserves the type detected from the prefix;
  contextual mappings affect only the resolved type and relation.
* Structured probes carry per-target verifier byte requirements,
  separate from their full discovery budgets.
* Dynamic filename rules can require both a prefix and suffix,
  limiting map group cluster matching to XML files.
* EDDS discovery recognizes the `ENF1` header marker
  while retaining legacy block-table detection.
* Thorough global prefix budgets include discovery probes only;
  verifier requirements remain path-specific.
* Registry traits mark confirmed archive/container formats
  and explicit source/compiled variants.
* Thorough model config pattern probes skip regexp scans when `CfgModels`
  is absent from the prefix.

## [0.2.1][] - 2026-04-03

### Added

* `aex` and `smd` file support

[0.2.1]: https://github.com/WoozyMasta/bimime/compare/v0.2.0...v0.2.1

## [0.2.0][] - 2026-04-01

### Added

* Plan helpers: `PlanFast`, `PlanNormal`, `PlanStrict`.
* Reusable analyzer API: `NewAnalyzer` and `Analyzer` methods.
* BI preset helpers for ambiguous RAP-like extensions:
  `BIAmbiguousRAPOverrides`, `BIAmbiguousRAPOptions`.

### Changed

* Breaking: API moved to plan-based configuration:
  `AnalyzeOptions` + `AnalyzePlan`
  (`AnalyzeMatchMode`, `AnalyzeValidateMode`).
* Breaking: `Analyze`, `AnalyzeReader`, `AnalyzeFile`, and
  `NeedsContent` now use unified `AnalyzeOptions`.
* Strict validation now includes content checks with explicit issues:
  `content_pattern_mismatch`, `insufficient_content`.
* Registry was expanded and refined (types/signatures/descriptions).

### Removed

* Breaking: legacy detect-mode API removed
  (`DetectMode*`, `ParseDetectMode`, `ErrInvalidDetectMode`).

[0.2.0]: https://github.com/WoozyMasta/bimime/compare/v0.1.1...v0.2.0

## [0.1.0][] - 2026-02-22

### Added

* First public release

[0.1.0]: https://github.com/WoozyMasta/bimime/tree/v0.1.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
