# bimime

`bimime` classifies BI game file types
from logical names and bounded content prefixes.
The package uses the same analysis rules for filesystem files,
readers, memory buffers, and archive entries.

## Profiles

Profiles set the evidence and I/O budget. They do not validate complete files.

* `ProfileFast` uses path evidence only and requires no content reads.
* `ProfileNormal` reads only for useful refinements of ambiguous BI types.
* `ProfileThorough` enables bounded content discovery and path-scoped patterns.

## Filesystem

Use `AnalyzeFile` when the package should open the path
and read its planned prefix:

```go
result, err := bimime.AnalyzeFile("terrain.wrp", bimime.ProfileNormal)
if err != nil {
    return err
}

fmt.Println(result.Type, result.Relation)
```

## Caller-owned storage

Use `RequiredPrefix` and `Analyze`
when the caller owns file, archive, or memory storage.
`RequiredPrefix` is the total prefix length from offset zero.

```go
size := bimime.RequiredPrefix(entry.Name, bimime.ProfileNormal)
prefix, err := entry.ReadPrefix(size)
if err != nil {
    return err
}

result := bimime.Analyze(entry.Name, prefix, bimime.ProfileNormal)
```

`ProfileFast` always returns zero.
`ProfileNormal` requests content for
ambiguous P3D, WRP, RVMat, and BISurf classifications.
`ProfileThorough` uses a bounded global binary prefix
and requests a larger sample only for registered path-scoped text refinements.

## Results and registry

`Result` keeps the final `TypeID`, path and content evidence, their relation,
the content probe source, planned prefix length, and completeness flags.
A conflict does not imply full-file validation;
consumers can decide how to handle `RelationConflict`.

`Lookup` and `Registry` return safe copies of type descriptors.
Consumers can key their own policy by `TypeID`
or use neutral `FamilyID` and `Trait` values:

```go
var compressionHint = map[bimime.TypeID]bool{
    "text.sqf": true,
    "audio.ogg": false,
}

_ = compressionHint[result.Type]
```

The package describes file identity and objective traits.
Consumer-specific policy such as compression decisions
belongs in the consuming package.
