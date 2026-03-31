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
