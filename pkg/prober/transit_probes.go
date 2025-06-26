package prober

import (
	"fmt"
	"sync"
)

type transitProbes struct {
	m map[uint64]int64 // index is the sequence number, value is the timestamp
	l sync.RWMutex
}

func (t *transitProbes) add(p *probe) {
	t.l.Lock()
	defer t.l.Unlock()
	t.m[p.SequenceNumber] = p.TimeStamp
}

func (t *transitProbes) remove(seq uint64) error {
	t.l.Lock()

	if _, ok := t.m[seq]; !ok {
		t.l.Unlock()
		return fmt.Errorf("Sequence number %d not found", seq)
	}

	delete(t.m, seq)
	t.l.Unlock()
	return nil
}

func (t *transitProbes) getLt(lt int64) map[uint64]struct{} {
	ret := make(map[uint64]struct{})
	t.l.RLock()
	defer t.l.RUnlock()

	for seq, ts := range t.m {
		if ts < lt {
			ret[seq] = struct{}{}
		}
	}

	return ret
}

func newTransitProbes() *transitProbes {
	return &transitProbes{
		m: make(map[uint64]int64),
	}
}
