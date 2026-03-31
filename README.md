# bimime

`bimime` is a Go package for detecting Bi ecosystem file types
by path hint, magic bytes, and lightweight text checks.
It covers Real Virtuality and Enfusion engine formats,
plus common modding files such as configs, scripts, assets,
localization tables, project/workbench files,
and diagnostics/crash artifacts.

## Detection Profiles

* fast: path/extension only, no content read.
* normal: extension + magic when needed by plan/extension.
* strict: normal + validation checks:
  extension/magic mismatch, text-likeness, content patterns.

## Advanced Plans

* `Analyze` supports per-extension `AnalyzePlan` overrides.
* Use `AnalyzeMatchExtensionMagic` for targeted forced magic probing
  on selected extensions (for example `wrp`, `p3d`, `rvmat`, `bisurf`).
* For batch processing, prefer `Analyzer` to reuse normalized config.

## Usage

```go
result, err := bimime.AnalyzeFile(
    bimime.BIAmbiguousRAPOptions("terrain.wrp", nil),
)
if err != nil {
    return err
}

fmt.Println(result.Probe.Resolved.ID)
```

```go
analyzer := bimime.NewAnalyzer(bimime.AnalyzeOptions{
    DefaultPlan:      bimime.PlanFast(),
    PlansByExtension: bimime.BIAmbiguousRAPOverrides(),
})

result, err := analyzer.AnalyzeFile("terrain.wrp")
if err != nil {
    return err
}

fmt.Println(result.Probe.Resolved.ID)
```

Equivalent explicit options:

```go
result, err := bimime.AnalyzeFile(bimime.AnalyzeOptions{
    Path: "terrain.wrp",
    DefaultPlan:      bimime.PlanFast(),
    PlansByExtension: bimime.BIAmbiguousRAPOverrides(),
})
if err != nil {
    return err
}

fmt.Println(result.Probe.Resolved.ID)
```

```go
result, err := bimime.AnalyzeFile(bimime.AnalyzeOptions{
    Path: "x.png",
    DefaultPlan: bimime.PlanNormal(),
})
if err != nil {
    return err
}

result = bimime.Analyze(bimime.AnalyzeOptions{
    Path:   "script.sqf",
    Prefix: dataPrefix,
    DefaultPlan: bimime.PlanNormal(),
    PlansByExtension: map[string]bimime.AnalyzePlan{
        "sqf": {
            Match:    bimime.AnalyzeMatchExtensionMagicNeeded,
            Validate: bimime.AnalyzeValidateStrict,
        },
    },
})
```

## Behavior Notes

* `NeedsContent` decides whether prefix bytes are required.
* `AnalyzeReader` reads only a prefix, not whole payload.
* `AnalyzeFile` opens file and reads only required prefix bytes.
* `Probe` resolves type by extension and magic bytes.
* `strict` without payload prefix returns `insufficient_content`.
