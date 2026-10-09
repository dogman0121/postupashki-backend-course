package semaphore

import (
	"primitives/internal/futex"
	"runtime"
	"sync/atomic"
)

type Semaphore struct {
	permits uint32
	waiters uint32
}

func New(n int) *Semaphore {
	if n < 0 {
		panic("Счетчик семафора не может быть натуральным числом")
	}

	return &Semaphore{permits: uint32(n), waiters: 0}
}

func (s *Semaphore) Acquire() {
	for {
		current := atomic.LoadUint32(&s.permits)

		if current == 0 {
			atomic.AddUint32(&s.waiters, 1)

			futex.Wait(&s.permits, 0)

			atomic.AddUint32(&s.waiters, ^uint32(0))
			continue
		}

		if atomic.CompareAndSwapUint32(&s.permits, current, current-1) {
			return
		}

		runtime.Gosched()
	}
}

func (s *Semaphore) TryAcquire() bool {
	for {
		current := atomic.LoadUint32(&s.permits)

		if current == 0 {
			return false
		}

		if atomic.CompareAndSwapUint32(&s.permits, current, current-1) {
			return true
		}

		runtime.Gosched()
	}
}

func (s *Semaphore) Release() {
	atomic.AddUint32(&s.permits, 1)

	if atomic.LoadUint32(&s.waiters) > 0 {
		futex.Wake(&s.permits)
	}
}

func (s *Semaphore) Available() int {
	return int(atomic.LoadUint32(&s.permits))
}
