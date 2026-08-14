// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/woozymasta/bimime"
)

func ExampleAnalyzeFile() {
	dir, err := os.MkdirTemp("", "bimime-example-")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	path := filepath.Join(dir, "model.p3d")
	if err := os.WriteFile(path, []byte("ODOL"), 0o600); err != nil {
		panic(err)
	}

	result, err := bimime.AnalyzeFile(path, bimime.ProfileNormal)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Type)
	// Output: bi.p3d.odol
}

func ExampleAnalyzeReader() {
	result, err := bimime.AnalyzeReader("model.p3d", strings.NewReader("ODOL"), bimime.ProfileNormal)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Type)
	// Output: bi.p3d.odol
}

func ExampleRequiredPrefix() {
	name := "model.p3d"
	data := []byte("ODOL")
	prefixSize := bimime.RequiredPrefix(name, bimime.ProfileNormal)
	result := bimime.Analyze(name, data[:prefixSize], bimime.ProfileNormal)

	fmt.Printf("read %d bytes: %s\n", prefixSize, result.Type)
	// Output: read 4 bytes: bi.p3d.odol
}

func ExampleProbe() {
	name := "scripts/misnamed.sqf"
	prefix := []byte{0x89, 'P', 'N', 'G', 13, 10, 26, 10}
	result := bimime.Probe(name, prefix)

	if result.Relation == bimime.RelationConflict {
		fmt.Printf("path=%s content=%s relation=conflict\n", result.ByPath, result.ByContent)
	}

	// Output:
	// path=text.sqf content=image.png relation=conflict
}
