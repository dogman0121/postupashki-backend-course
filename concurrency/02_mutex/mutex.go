package mutex

import (
	"primitives/internal/futex"
	"runtime"
	"sync/atomic"
)

const (
	free      = iota // Мьютекс свободен
	held             // Мьютекс занят, никто не ждет
	contended        // Мьютекс занят, кто-то ждет
)

const maxWaitingIterations = 4

type Mutex struct {
	state uint32
}

func (m *Mutex) Lock() {
	// крутимся, ждем пока освободится TTAS если никто не ждет
	iterations := 0
	for iterations < maxWaitingIterations {
		val := atomic.LoadUint32(&m.state)
		if val == free && atomic.CompareAndSwapUint32(&m.state, free, held) {
			return
		}
		iterations += 1

		runtime.Gosched()
	}

	for {

		c := atomic.LoadUint32(&m.state)

		if c == contended || atomic.CompareAndSwapUint32(&m.state, held, contended) {
			futex.Wait(&m.state, contended)
		}

		if atomic.CompareAndSwapUint32(&m.state, free, contended) {
			return
		}

		runtime.Gosched()
	}
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapUint32(&m.state, free, held)
}

func (m *Mutex) Unlock() {
	old := atomic.SwapUint32(&m.state, free)

	if old == free {
		panic("Mutex был захвачен")
	}

	if old == contended {
		futex.Wake(&m.state)
	}
}
