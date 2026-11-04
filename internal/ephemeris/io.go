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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func textBytes(s string) []byte {
	return append([]byte(s), 0)
}

func textString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b := v.([]byte)
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return string(b)
}

func copyText(dst []byte, src any) []byte {
	s := textString(src)
	n := copy(dst, s)
	if n < len(dst) {
		dst[n] = 0
	}
	return dst
}

func copyTextN(dst []byte, src any, count any) []byte {
	n := int(integer(count))
	if n > len(dst) {
		n = len(dst)
	}
	s := textString(src)
	for i := 0; i < n; i++ {
		if i < len(s) {
			dst[i] = s[i]
		} else {
			dst[i] = 0
		}
	}
	return dst
}

func appendText(dst []byte, src any) []byte {
	n := int(textLength(dst))
	if n < len(dst) {
		copyText(dst[n:], src)
	}
	return dst
}

func appendTextN(dst []byte, src any, count any) []byte {
	s := textString(src)
	n := int(integer(count))
	if n < len(s) {
		s = s[:n]
	}
	return appendText(dst, s)
}

func textLength(v any) uint64 {
	return uint64(len(textString(v)))
}

func compareText(a, b any) int32 {
	return int32(strings.Compare(textString(a), textString(b)))
}

func compareTextN(a, b any, n any) int32 {
	x, y := textString(a), textString(b)
	c := int(integer(n))
	if len(x) > c {
		x = x[:c]
	}
	if len(y) > c {
		y = y[:c]
	}
	return int32(strings.Compare(x, y))
}

func findByte(s []byte, c any) []byte {
	i := bytes.IndexByte(s[:min(len(s), int(textLength(s))+1)], byte(integer(c)))
	if i < 0 {
		return nil
	}
	return s[i:]
}

func findLastByte(s []byte, c any) []byte {
	i := bytes.LastIndexByte(s[:min(len(s), int(textLength(s))+1)], byte(integer(c)))
	if i < 0 {
		return nil
	}
	return s[i:]
}

func findText(s []byte, sub any) []byte {
	i := strings.Index(textString(s), textString(sub))
	if i < 0 {
		return nil
	}
	return s[i:]
}

func formatText(dst []byte, format any, args ...any) int32 {
	for i, v := range args {
		if _, ok := v.([]byte); ok {
			args[i] = textString(v)
		}
	}
	f := textString(format)
	f = strings.ReplaceAll(f, "%ld", "%d")
	f = strings.ReplaceAll(f, "%lu", "%d")
	s := fmt.Sprintf(f, args...)
	copyText(dst, s)
	return int32(len(s))
}

func formatTextN(dst []byte, count any, format any, args ...any) int32 {
	n := min(len(dst), int(integer(count)))
	return formatText(dst[:n], format, args...)
}

func integer(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case uint8:
		return int64(n)
	case uint16:
		return int64(n)
	case uint32:
		return int64(n)
	case uint64:
		return int64(n)
	case float64:
		return int64(n)
	default:
		panic("ephemeris: unsupported numeric argument")
	}
}

func upperByte(v any) int32 {
	c := int32(integer(v))
	if c >= 'a' && c <= 'z' {
		c -= 'a' - 'A'
	}
	return c
}

func lowerByte(v any) int32 {
	c := int32(integer(v))
	if c >= 'A' && c <= 'Z' {
		c += 'a' - 'A'
	}
	return c
}

func isSpace(v any) int32 {
	if strings.ContainsRune(" \t\n\r\v\f", rune(integer(v))) {
		return 1
	}
	return 0
}

func isDigit(v any) int32 {
	c := integer(v)
	if c >= '0' && c <= '9' {
		return 1
	}
	return 0
}

func parseInt(v any) int32 {
	return int32(parseLong(v))
}

func parseLong(v any) int64 {
	s := strings.TrimLeft(textString(v), " \t\n\r\v\f")
	i := 0
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		i++
	}
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	n, _ := strconv.ParseInt(s[:i], 10, 64)
	return n
}

func boolInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func absInt(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

func environmentText(name any) []byte {
	s, ok := os.LookupEnv(textString(name))
	if !ok {
		return nil
	}
	return textBytes(s)
}

// sliceOffset returns the element offset between suffixes of the same slice.
func sliceOffset[T any](a, b []T) int32 {
	return int32(len(b) - len(a))
}

type textFile struct {
	file   *os.File
	reader *bufio.Reader
	err    error
}

func openTextFile(path, mode any) *textFile {
	f, err := os.Open(textString(path))
	if err != nil {
		return nil
	}
	return &textFile{file: f, reader: bufio.NewReader(f)}
}

func closeTextFile(f *textFile) int32 {
	if f == nil {
		return -1
	}
	if err := f.file.Close(); err != nil {
		return -1
	}
	return 0
}

func readLine(dst []byte, size any, f *textFile) []byte {
	if f == nil {
		return nil
	}
	limit := min(len(dst), int(integer(size))) - 1
	n := 0
	for n < limit {
		value, err := f.reader.ReadByte()
		if err == nil {
			dst[n] = value
			n++
			if value == '\n' {
				break
			}
		}
		if err != nil {
			f.err = err
			break
		}
	}
	if n < len(dst) {
		dst[n] = 0
	}
	if n == 0 {
		return nil
	}
	return dst
}

func seekFile(f *textFile, offset, whence any) int32 {
	if f == nil {
		return -1
	}
	distance, origin := integer(offset), int(integer(whence))
	if origin == io.SeekCurrent {
		distance -= int64(f.reader.Buffered())
	}
	_, err := f.file.Seek(distance, origin)
	f.reader.Reset(f.file)
	f.err = err
	if err != nil {
		return -1
	}
	return 0
}

func tellFile(f *textFile) int64 {
	if f == nil {
		return -1
	}
	n, err := f.file.Seek(0, io.SeekCurrent)
	f.err = err
	if err != nil {
		return -1
	}
	return n - int64(f.reader.Buffered())
}

func rewindFile(f *textFile) {
	seekFile(f, 0, 0)
}

func readBinary(dst any, size, count any, f *textFile) uint64 {
	if f == nil {
		return 0
	}
	n, c := int(integer(size)), int(integer(count))
	if n <= 0 || c <= 0 || n > numericCapacity(dst)/c {
		return 0
	}
	data := make([]byte, n*c)
	read, err := io.ReadFull(f.reader, data)
	f.err = err
	decodeBinary(dst, data[:read])
	return uint64(read / n)
}
