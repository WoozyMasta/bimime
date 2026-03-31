// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import "errors"

var (
	// ErrNilReader is returned when AnalyzeReader receives nil reader.
	ErrNilReader = errors.New("nil reader")
)
