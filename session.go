// SPDX-License-Identifier: AGPL-3.0-only
package swisseph

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"modernc.org/libc"
	"modernc.org/libc/sys/types"
)

// Session exposes the complete public C API with checked Go buffers. Obtain it
// through WithSession. C return codes and serr buffers are preserved; a Go error
// means invalid buffers, non-finite input, a closed session or allocation failure.
// Never call Engine methods or nest WithSession inside its callback: the callback
// already owns the global engine lock. Use the Session methods throughout.
// Buffers are copied into independent native-layout allocations. Do not rely on
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
	tls  *libc.TLS
	ptrs []uintptr
	back []func()
	err  error
}

func (f *frame) alloc(size int) uintptr {
	if f.err != nil {
		return 0
	}
	p := libc.Xcalloc(f.tls, 1, types.Size_t(size))
	if p == 0 {
		f.err = errors.New("ephemeris allocation failed")
		return 0
	}
	f.ptrs = append(f.ptrs, p)
	return p
}
func (f *frame) close() {
	for _, p := range f.ptrs {
		libc.Xfree(f.tls, p)
	}
}
func (f *frame) copyBack() {
	for _, fn := range f.back {
		fn()
	}
}
func (f *frame) length(name string, got, min int) bool {
	if got < min {
		f.err = fmt.Errorf("%s buffer length %d, require at least %d", name, got, min)
		return false
	}
	return f.err == nil
}
func (f *frame) text(name, value string, max int) uintptr {
	if len(value) > max || strings.IndexByte(value, 0) >= 0 {
		f.err = fmt.Errorf("%s must be NUL-free and at most %d bytes", name, max)
		return 0
	}
	p := f.alloc(len(value) + 1)
	if p != 0 {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(p)), len(value)+1), value)
	}
	return p
}
func (f *frame) bytes(name string, value []byte, min int) uintptr {
	if !f.length(name, len(value), min) {
		return 0
	}
	// C string arguments must be terminated within the checked buffer.
	terminated := false
	for _, v := range value {
		if v == 0 {
			terminated = true
			break
		}
	}
	if !terminated {
		f.err = fmt.Errorf("%s requires a NUL terminator", name)
		return 0
	}
	p := f.alloc(len(value))
	if p == 0 {
		return 0
	}
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(p)), len(value))
	copy(buffer, value)
	f.back = append(f.back, func() { copy(value, buffer) })
	return p
}
func (f *frame) doubles(name string, value []float64, min int) uintptr {
	if !f.length(name, len(value), min) {
		return 0
	}
	if !finite(value...) {
		f.err = invalidNumber(name)
		return 0
	}
	p := f.alloc(len(value) * 8)
	if p == 0 {
		return 0
	}
	buffer := unsafe.Slice((*float64)(unsafe.Pointer(p)), len(value))
	copy(buffer, value)
	f.back = append(f.back, func() { copy(value, buffer) })
	return p
}
func (f *frame) ints(name string, value []int32, min int) uintptr {
	if !f.length(name, len(value), min) {
		return 0
	}
	p := f.alloc(len(value) * 4)
	if p == 0 {
		return 0
	}
	buffer := unsafe.Slice((*int32)(unsafe.Pointer(p)), len(value))
	copy(buffer, value)
	f.back = append(f.back, func() { copy(value, buffer) })
	return p
}
