// Package mutex provides a context aware mutex.
//
// Example:
//
//		var m = mutex.New()
//		ctx, cancel := context.WithCancel(context.Background())
//		go func() {
//			<-time.After(time.Minute)
//			cancel()
//		}()
//
//		if !m.TryLock(ctx) {
//			return
//		}
//		defer m.Unlock()
//
//
package mutex

import (
	"context"
	"sync/atomic"
)

// Mutex is a context aware mutex
type Mutex struct {
	wl chan struct{}
	l  int32
}

// New returns a new mutex
func New() *Mutex {
	m := &Mutex{
		wl: make(chan struct{}, 1),
	}

	// we always start unlocked
	m.wl <- struct{}{}

	return m
}

// TryLock blocks until the provided context is canceled or the lock succeeded
// The caller must check the return value of TryLock() to ensure the mutex got
// locked.
func (m *Mutex) TryLock(ctx context.Context) bool {
	select {
	case <-m.wl:
		atomic.StoreInt32(&m.l, 1)
		return true
	case <-ctx.Done():
		return false
	}
}

// Lock locks the mutex and blocks until the lock succeeds. It will never return
// without the mutex being locked. This provides compatability to the standard
// sync.Mutex
func (m *Mutex) Lock() {
	if ok := m.TryLock(context.Background()); ok == false {
		// that must work!
		panic("failed to acquire lock")
	}
}

// Unlock unlocks the mutex. It must be called after a successful call to TryLock()
// or Lock()
func (m *Mutex) Unlock() {
	atomic.StoreInt32(&m.l, 0)
	m.wl <- struct{}{}
}

// IsLocked returns true if the mutex is currently locked
// Do **not** use this function to decide if an operation is safe or not! It may become
// locked/unlocked right after you checked it
func (m *Mutex) IsLocked() bool {
	return atomic.LoadInt32(&m.l) == 1
}
