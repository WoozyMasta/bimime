# bimime

[bimime][] is a Go library for classifying BI game files
from logical names and bounded content probes.
It reports path and content evidence
without claiming to validate the complete file.

## Features

* Fast, Normal, and Thorough profiles with bounded read requirements.
* Classification from paths, readers, files, and caller-owned data.
* Separate path and content evidence, including conflicts and refinements.
* Stable type identifiers and neutral traits for consumer-owned policy.

## Installation

```sh
go get github.com/woozymasta/bimime@latest
```

Requires Go 1.21 or later.

## Documentation

See the [API documentation][bimime] for package usage.
Runnable examples are maintained in [`example_test.go`](example_test.go).

<!-- links -->

[bimime]: https://pkg.go.dev/github.com/woozymasta/bimime
