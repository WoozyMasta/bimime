// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import "errors"

var (
	// ErrNilReader is returned when AnalyzeReader needs content from a nil reader.
	ErrNilReader = errors.New("nil reader")

	// errInvalidRegistry marks invalid built-in registry declarations.
	errInvalidRegistry = errors.New("invalid registry")
)
