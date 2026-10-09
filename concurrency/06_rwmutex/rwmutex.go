package rwmutex

import (
	"primitives/internal/futex"
	"runtime"
	"sync/atomic"
)

const writer = 1 << 31

type RWMutex struct {
	state           uint32
	writers_waiting atomic.Bool
}

func (rw *RWMutex) RLock() { // Блокирует на чтение
	for {
		curr := atomic.LoadUint32(&rw.state)

		if rw.writers_waiting.Load() {
			runtime.Gosched()
			continue
		}

		if curr >= writer {
			futex.Wait(&rw.state, curr)
			continue
		}

		if atomic.CompareAndSwapUint32(&rw.state, curr, curr+1) {
			return // Успешно захватили блокировку на чтение
		}

		runtime.Gosched()
	}
}

func (rw *RWMutex) RUnlock() { // Разблокирует на чтение
	for {
		old := atomic.LoadUint32(&rw.state)

		if old == 0 {
			panic("RWMutex не заблокирован")
		}

		if atomic.CompareAndSwapUint32(&rw.state, old, old-1) {
			if old-1 == 0 {
				futex.Wake(&rw.state)
			}
			return
		}

		runtime.Gosched()
	}
}

func (rw *RWMutex) Lock() { // Блокирует на запись
	rw.writers_waiting.Store(true)

	for {
		curr := atomic.LoadUint32(&rw.state)

		if curr == 0 {
			if atomic.CompareAndSwapUint32(&rw.state, 0, writer) {
				rw.writers_waiting.Store(false)
				return
			}
			continue
		}

		futex.Wait(&rw.state, curr)
	}
}

func (rw *RWMutex) Unlock() { // Блокирует на запись
	if !atomic.CompareAndSwapUint32(&rw.state, writer, 0) {
		panic("RWMutex не заблокирован")
	}

	futex.WakeAll(&rw.state)
}
