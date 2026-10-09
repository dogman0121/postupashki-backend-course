package waitgroup

import (
	"primitives/internal/futex"
	"runtime"
	"sync/atomic"
)

type WaitGroup struct {
	count uint32
}

func (wg *WaitGroup) Add(delta int) {
	for {
		old := atomic.LoadUint32(&wg.count)

		v := int(old) + delta

		if v < 0 {
			panic("Счетчик не может быть меньше нуля")
		}

		if atomic.CompareAndSwapUint32(&wg.count, old, uint32(v)) {
			if v == 0 {
				futex.WakeAll(&wg.count)
			}

			return
		}

		runtime.Gosched()
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	for {
		curr := atomic.LoadUint32(&wg.count)
		if curr == 0 {
			return
		}

		futex.Wait(&wg.count, curr)

		runtime.Gosched()
	}
}
