//go:build (darwin && (amd64 || arm64)) || (linux && amd64)

// SPDX-License-Identifier: AGPL-3.0-only
package swisseph

// Plain C char is signed on these targets.
type abiChar = int8
