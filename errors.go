// SPDX-License-Identifier: MIT
// Copyright (c) 2026 WoozyMasta
// Source: github.com/woozymasta/bimime

package bimime

import "errors"

var (
	// ErrInvalidDetectMode is returned for unsupported detect mode values.
	ErrInvalidDetectMode = errors.New("invalid detect mode")
	// ErrNilReader is returned when AnalyzeReader receives nil reader.
	ErrNilReader = errors.New("nil reader")
)
