package spinlock

import (
	"runtime"
	"sync/atomic"
)

type Spinlock struct {
	locked atomic.Bool
}

func (s *Spinlock) Lock() {
	for !s.locked.CompareAndSwap(false, true) {
		runtime.Gosched()
	}
}

func (s *Spinlock) TryLock() bool {
	return s.locked.CompareAndSwap(false, true)
}

func (s *Spinlock) Unlock() {
	sharedUnlock(&s.locked)
}

type TTAS struct {
	locked atomic.Bool
}

func (s *TTAS) Lock() {
	for {
		locked := s.locked.Load()

		if !locked && s.locked.CompareAndSwap(false, true) {
			return
		}

		runtime.Gosched()
	}
}

func (s *TTAS) TryLock() bool {
	return s.locked.CompareAndSwap(false, true)
}

func (s *TTAS) Unlock() {
	sharedUnlock(&s.locked)
}

func sharedUnlock(atomic *atomic.Bool) {
	if !atomic.CompareAndSwap(true, false) {
		panic("Spinlock не был захвачен")
	}
}
