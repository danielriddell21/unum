// Package shared provides static assets shared across all unum web UIs.
package shared

import "embed"

//go:embed assets/*
var Assets embed.FS
