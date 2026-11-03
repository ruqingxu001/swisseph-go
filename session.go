// SPDX-License-Identifier: AGPL-3.0-only
package swisseph

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Session exposes the complete public C API with checked Go buffers. Obtain it
// through WithSession. Swiss Ephemeris return codes and serr buffers are preserved; a Go error
// means invalid buffers, non-finite input or a closed session. Each callback owns
// its calculation state; nested Engine operations use a separate state.
// Buffers are copied into independent Go slices. Do not rely on
// overlapping buffers aliasing as they would in C. Callers own buffer synchronization.
// Methods preserve C domain rules; this is not a validator for arbitrary C inputs.
type Session struct {
	mu     sync.Mutex
	c      *call
	active bool
}

// WithSession keeps settings, file caches and operations in one isolated scope.
// The session becomes unusable when fn returns, including when fn panics.
func (e *Engine) WithSession(fn func(*Session) error) error {
	if fn == nil {
		return errors.New("nil session callback")
	}
	return e.run(func(c *call) error {
		s := &Session{c: c, active: true}
		defer func() { s.mu.Lock(); s.active = false; s.mu.Unlock() }()
		return fn(s)
	})
}
func (s *Session) enter() error {
	if s == nil {
		return errors.New("nil ephemeris session")
	}
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return errors.New("ephemeris session is closed")
	}
	return nil
}
func invalidNumber(name string) error { return fmt.Errorf("%s must be finite", name) }

type frame struct {
	back []func()
	err  error
}

func (f *frame) copyBack() {
	for _, copy := range f.back {
		copy()
	}
}
func (f *frame) length(name string, got, min int) bool {
	if got < min {
		f.err = fmt.Errorf("%s buffer length %d, require at least %d", name, got, min)
		return false
	}
	return f.err == nil
}
func (f *frame) text(name, value string, max int) string {
	if len(value) > max || strings.IndexByte(value, 0) >= 0 {
		f.err = fmt.Errorf("%s must be NUL-free and at most %d bytes", name, max)
	}
	return value
}
func (f *frame) bytes(name string, value []byte, min int) []byte {
	if !f.length(name, len(value), min) {
		return nil
	}
	if !bytes.Contains(value, []byte{0}) {
		f.err = fmt.Errorf("%s requires a NUL terminator", name)
		return nil
	}
	buffer := append([]byte(nil), value...)
	f.back = append(f.back, func() { copy(value, buffer) })
	return buffer
}
func (f *frame) doubles(name string, value []float64, min int) []float64 {
	if !f.length(name, len(value), min) {
		return nil
	}
	if !finite(value...) {
		f.err = invalidNumber(name)
		return nil
	}
	buffer := append([]float64(nil), value...)
	f.back = append(f.back, func() { copy(value, buffer) })
	return buffer
}
func (f *frame) ints(name string, value []int32, min int) []int32 {
	if !f.length(name, len(value), min) {
		return nil
	}
	buffer := append([]int32(nil), value...)
	f.back = append(f.back, func() { copy(value, buffer) })
	return buffer
}
