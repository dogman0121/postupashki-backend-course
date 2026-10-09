package barrier

import (
	"primitives/internal/futex"
	"sync/atomic"
)

type Barrier struct {
	need    uint32
	arrived uint32
	round   uint32
}

func New(n int) *Barrier {
	if n <= 1 {
		panic("Barrier не может быть отрицательным")
	}
	return &Barrier{need: uint32(n), arrived: 0, round: 0}
}

func (b *Barrier) Wait() {
	round := atomic.LoadUint32(&b.round)
	for {
		old := atomic.AddUint32(&b.arrived, 1)

		if old == b.need {
			atomic.SwapUint32(&b.arrived, 0)
			atomic.AddUint32(&b.round, 1)

			futex.WakeAll(&b.round)
			return
		}

		futex.Wait(&b.round, round)

		break
	}
}
