// SPDX-License-Identifier: AGPL-3.0-only
package raw

import (
	"math"
	"modernc.org/libc"
	"strconv"
	"strings"
)

// cAtof implements the C-locale numeric-prefix semantics Swiss Ephemeris needs.
// libc v1.62.1 aborts on strings such as "ICRS"; C atof returns zero instead.
// This adapter changes no astronomical formula and is applied by generate.py.
func cAtof(_ *libc.TLS, p uintptr) float64 { return parseCFloat(libc.GoString(p)) }
func parseCFloat(s string) float64 {
	s = strings.TrimLeft(s, " \t\n\r\v\f")
	if s == "" {
		return 0
	}
	start := 0
	sign := 1.0
	if s[0] == '+' || s[0] == '-' {
		if s[0] == '-' {
			sign = -1
		}
		start++
	}
	lower := strings.ToLower(s[start:])
	if strings.HasPrefix(lower, "inf") {
		return math.Inf(int(sign))
	}
	if strings.HasPrefix(lower, "nan") {
		return math.Copysign(math.NaN(), sign)
	}
	i := start
	hex := false
	if len(s) >= i+2 && s[i] == '0' && (s[i+1] == 'x' || s[i+1] == 'X') {
		hex = true
		i += 2
	}
	digit := func(c byte) bool {
		return c >= '0' && c <= '9' || hex && (c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F')
	}
	digits := 0
	for i < len(s) && digit(s[i]) {
		i++
		digits++
	}
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && digit(s[i]) {
			i++
			digits++
		}
	}
	if digits == 0 {
		if hex {
			return math.Copysign(0, sign)
		}
		return 0
	}
	exponent := false
	if i < len(s) && ((!hex && (s[i] == 'e' || s[i] == 'E')) || (hex && (s[i] == 'p' || s[i] == 'P'))) {
		before := i
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		j := i
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i == j {
			i = before
		} else {
			exponent = true
		}
	}
	token := s[:i]
	if hex && !exponent {
		token += "p0"
	}
	value, _ := strconv.ParseFloat(token, 64)
	return value
}
