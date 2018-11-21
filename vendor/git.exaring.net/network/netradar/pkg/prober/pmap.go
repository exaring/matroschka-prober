package prober

import (
	"fmt"
	"sync"
)

type pMap struct {
	m map[uint64]int64
	l sync.RWMutex
}

func (pm *pMap) add(p *probe) {
	pm.l.Lock()
	defer pm.l.Unlock()
	pm.m[p.Seq] = p.Ts
}

func (pm *pMap) remove(s uint64) error {
	pm.l.Lock()
	defer pm.l.Unlock()
	if _, ok := pm.m[s]; !ok {
		return fmt.Errorf("Sequence number %d not found", s)
	}
	delete(pm.m, s)
	return nil
}

func (pm *pMap) getLt(lt int64) map[uint64]int64 {
	ret := make(map[uint64]int64)
	pm.l.RLock()
	defer pm.l.RUnlock()

	for s, ts := range pm.m {
		if ts < lt {
			ret[s] = ts
		}
	}

	return ret
}

func newpMap() *pMap {
	return &pMap{
		m: make(map[uint64]int64),
	}
}
