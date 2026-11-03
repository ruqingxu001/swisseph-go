// SPDX-License-Identifier: AGPL-3.0-only
// Derived from Swiss Ephemeris 2.10.03; see NOTICE.md and upstream/LICENSE.
// Modified for native Go records, slices and isolated state, 2026-11-03.

/* Copyright (C) 1997 - 2021 Astrodienst AG, Switzerland.  All rights reserved.

License conditions
------------------

This file is part of Swiss Ephemeris.

Swiss Ephemeris is distributed with NO WARRANTY OF ANY KIND.  No author
or distributor accepts any responsibility for the consequences of using it,
or for whether it serves any particular purpose or works at all, unless he
or she says so in writing.

Swiss Ephemeris is made available by its authors under a dual licensing
system. The software developer, who uses any part of Swiss Ephemeris
in his or her software, must choose between one of the two license models,
which are
a) GNU Affero General Public License (AGPL)
b) Swiss Ephemeris Professional License

The choice must be made before the software developer distributes software
containing parts of Swiss Ephemeris to others, and before any public
service using the developed software is activated.

If the developer choses the AGPL software license, he or she must fulfill
the conditions of that license, which includes the obligation to place his
or her whole software project under the AGPL or a compatible license.
See https://www.gnu.org/licenses/agpl-3.0.html

If the developer choses the Swiss Ephemeris Professional license,
he must follow the instructions as found in http://www.astro.com/swisseph/
and purchase the Swiss Ephemeris Professional Edition from Astrodienst
and sign the corresponding license contract.

The License grants you the right to use, copy, modify and redistribute
Swiss Ephemeris, but only under certain conditions described in the License.
Among other things, the License requires that the copyright notices and
this notice be preserved on all copies.

Authors of the Swiss Ephemeris: Dieter Koch and Alois Treindl

The authors of Swiss Ephemeris have no control or influence over any of
the derived works, i.e. over software or services created by other
programmers which use Swiss Ephemeris functions.

The names of the authors or of the copyright holder (Astrodienst) must not
be used for promoting any software, product or service which uses or contains
the Swiss Ephemeris. This copyright notice is the ONLY place where the
names of the authors can legally appear, except in cases where they have
given special permission in writing.
*/
package ephemeris

import (
	"math"
	"strconv"
	"strings"
)

func parseFloat(value any) float64 {
	return parseCFloat(textString(value))
}

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
	if i < len(s) && (!hex && (s[i] == 'e' || s[i] == 'E') || hex && (s[i] == 'p' || s[i] == 'P')) {
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
