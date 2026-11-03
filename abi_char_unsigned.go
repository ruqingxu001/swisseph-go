//go:build linux && arm64

// SPDX-License-Identifier: AGPL-3.0-only
package swisseph

// AArch64 Linux uses unsigned plain C char. Preserve the caller's eight bits.
type abiChar = uint8
