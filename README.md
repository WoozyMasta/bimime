# bimime

`bimime` is a Go package for detecting Bi ecosystem file types
by path hint, magic bytes, and lightweight text checks.
It covers Real Virtuality and Enfusion engine formats,
plus common modding files such as configs, scripts, assets,
localization tables, project/workbench files,
and diagnostics/crash artifacts.

## Detection Modes

* `fast` uses filename and extension only.
* `normal` uses extension and magic when content is needed.
* `strict` uses normal mode plus validation checks.
* strict checks include extension/magic mismatch detection
  for signature-based formats.
* strict checks include quick text-likeness validation for text-like types.

## Usage

```go
mode, err := bimime.ParseDetectMode("strict")
if err != nil {
    return err
}

result, err := bimime.AnalyzeFile(
    "config.rvmat",
    bimime.AnalyzeOptions{Mode: mode},
)
if err != nil {
    return err
}

fmt.Println(result.Probe.Resolved.ID)
fmt.Println(result.Probe.Resolved.Description)
fmt.Println(result.Valid)
```

## Behavior Notes

* `NeedsContent` lets caller skip file reads in fast paths.
* `AnalyzeReader` reads only a prefix buffer, not full file payload.
* `AnalyzeFile` opens file and reads only required prefix bytes.
* `Probe` resolves type by extension and magic bytes.
