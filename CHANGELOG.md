<!-- markdownlint-disable MD024 -->
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

## [0.3.0][] - 2026-04-03

### Added

* Recognize modern Enfusion EDDS headers through the `ENF1` marker,
  while retaining legacy EDDS detection.

### Changed

* Breaking: replace the plan/options and strict-validation API with
  `ProfileFast`, `ProfileNormal`, and `ProfileThorough` bounded analysis.
  Results now retain separate path and content evidence, their relation,
  and probe status. `RequiredPrefix` replaces `NeedsContent`
  with an exact byte budget; the reusable `Analyzer` API is removed.
* Lower the minimum supported Go version to 1.21.
* Registry metadata now identifies source, compiled, container,
  and archive formats for downstream policy.
* Speed up Thorough model-config checks when `CfgModels` is absent.

### Fixed

* Apply the planned prefix limit to caller-provided data,
  keeping `Analyze` consistent with reader and file analysis.
* Restrict weak PAA, BMP, and raw MP3 frame signatures to matching paths;
  they no longer discover unrelated binary files.
* Base missing-content flags on each verifier's required bytes.
  Optional `.cfg` and `.xml` refinements
  no longer report their marker as mandatory,
  and path-only verifiers do not inflate the global Thorough budget.
* Preserve raw content evidence in `ByContent`;
  contextual RAP mappings affect only the resolved type and relation.
* Match `mapgroupcluster` refinements only for names ending in `.xml`.

[0.3.0]: https://github.com/WoozyMasta/bimime/compare/v0.2.1...v0.3.0

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
