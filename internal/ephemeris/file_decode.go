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
	"fmt"
	"io"
	"sort"
	"strings"
)

func sameSlice[T any](a, b []T) bool {
	if len(a) == 0 || len(b) == 0 {
		return len(a) == 0 && len(b) == 0
	}
	return &a[0] == &b[0]
}

func findAnyText(s []byte, set any) []byte {
	for i, c := range s {
		if c == 0 {
			break
		}
		if strings.IndexByte(textString(set), c) >= 0 {
			return s[i:]
		}
	}
	return nil
}

func duplicateText(s any) []byte {
	return textBytes(textString(s))
}

func isAlphaNumeric(value any) int32 {
	c := integer(value)
	return boolInt(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z')
}

func printText(format any, args ...any) int32 {
	for i, v := range args {
		if _, ok := v.([]byte); ok {
			args[i] = textString(v)
		}
	}
	n, _ := fmt.Printf(textString(format), args...)
	return int32(n)
}

// swiRightTrim removes trailing ASCII whitespace from a NUL-terminated buffer.
func (state *State) swiRightTrim(s []byte) []byte {
	end := int(textLength(s))
	for end > 0 && isSpace(s[end-1]) != 0 {
		end--
		s[end] = 0
	}
	return s
}

func (state *State) doFread(target any, size, count, width int32, file *textFile, position int32, reorder, endian, fileIndex int32, message []byte) int32 {
	if file == nil || size <= 0 || size > width || width > 8 || count < 0 || count > int32(numericCapacity(target))/width {
		copyText(message, "Ephemeris file is damaged (1). ")
		return -1
	}
	if position >= 0 {
		seekFile(file, position, 0)
	}
	data := make([]byte, int(size*count))
	_, err := io.ReadFull(file.reader, data)
	file.err = err
	if err != nil {
		stage := 1
		if reorder != 0 || size != width {
			stage = 3
		}
		formatText(message, "Ephemeris file is damaged (%d). ", stage)
		name := textString(state.positions.data.files[fileIndex].name[:])
		if int(textLength(message))+len(name) < 255 {
			formatText(message, "Ephemeris file %s is damaged (%d).", name, stage+1)
		}
		return -1
	}
	values := make([]byte, int(count*width))
	for i := int32(0); i < count; i++ {
		for j := int32(0); j < size; j++ {
			dest := j
			if reorder != 0 {
				dest = size - j - 1
			}
			values[i*width+dest] = data[i*size+j]
		}
	}
	decodeBinary(target, values)
	return 0
}

func (state *State) saveStarInStruct(count int32, star *fixedStar, message []byte) int32 {
	if count < 1 {
		return -1
	}
	if len(state.positions.data.fixedStars) < int(count) {
		state.positions.data.fixedStars = append(state.positions.data.fixedStars, make([]fixedStar, int(count)-len(state.positions.data.fixedStars))...)
	}
	state.positions.data.fixedStars[count-1] = *star
	return 0
}

func fixedstarNameCompare(a, b any) int32 {
	x, y := a.(*fixedStar), b.(*fixedStar)
	return compareText(x.searchKey[:], y.searchKey[:])
}

func fstarNodeCompare(a, b any) int32 {
	return fixedstarNameCompare(a, b)
}

func sortRecords(stars []fixedStar, count, size any, compare func(any, any) int32) {
	n := int(integer(count))
	sort.Slice(stars[:n], func(i, j int) bool {
		return compare(&stars[i], &stars[j]) < 0
	})
}

func searchRecords(keyValue any, stars []fixedStar, count, size any, compare func(any, any) int32) *fixedStar {
	key, ok := keyValue.(*fixedStar)
	if !ok {
		key = &fixedStar{}
		copyText(key.searchKey[:], keyValue)
	}
	n := int(integer(count))
	i := sort.Search(n, func(i int) bool {
		return compare(key, &stars[i]) <= 0
	})
	if i < n && compare(key, &stars[i]) == 0 {
		return &stars[i]
	}
	return nil
}

func trimBufferEnd(buffer []byte, cutset string) {
	end := int(textLength(buffer))
	for end > 0 && strings.IndexByte(cutset, buffer[end-1]) >= 0 {
		end--
		buffer[end] = 0
	}
}

func reverse32(v uint32) uint32 {
	return v>>24 | v>>8&0x0000ff00 | v<<8&0x00ff0000 | v<<24
}
