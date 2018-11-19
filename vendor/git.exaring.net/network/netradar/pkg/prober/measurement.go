package prober

import (
	"sync"
	"time"

	"github.com/golang/glog"
)

// Measurement represents a measurement over a deltaT time window
type Measurement struct {
	Sent     uint64
	Received uint64
	RttSum   int64
	RttMin   int64
	RttMax   int64
	Rtts     []int64
}

type measurements struct {
	m map[int64]*Measurement
	l sync.RWMutex
}

func newMeasurements() *measurements {
	return &measurements{
		m: make(map[int64]*Measurement),
	}
}

func (m *measurements) addSent(ts int64) {
	m.l.Lock()
	defer m.l.Unlock()
	if m.m[ts] == nil {
		m.m[ts] = &Measurement{
			Rtts: make([]int64, 0),
		}
	}
	m.m[ts].Sent++
}

func (m *measurements) addRecv(sentTs int64, rtt int64, deltaT int64) {
	m.l.RLock()
	defer m.l.RUnlock()
	allignedTs := sentTs - (sentTs % deltaT)
	if _, ok := m.m[allignedTs]; !ok {
		glog.Errorf("Received probe at %d sent at %d with rtt %d after bucket %d was removed. Now=%d", sentTs+rtt, sentTs, allignedTs, rtt, time.Now().UnixNano())
		return
	}
	m.m[allignedTs].Rtts = append(m.m[allignedTs].Rtts, rtt)
	m.m[allignedTs].Received++
	m.m[allignedTs].RttSum += rtt
	if rtt < m.m[allignedTs].RttMin || m.m[allignedTs].RttMin == 0 {
		m.m[allignedTs].RttMin = rtt
	}
	if rtt > m.m[allignedTs].RttMax {
		m.m[allignedTs].RttMax = rtt
	}
}

func (m *measurements) removeOlder(ts int64) {
	m.l.Lock()
	defer m.l.Unlock()

	for t := range m.m {
		if t < ts {
			delete(m.m, t)
		}
	}
}

func (m *measurements) get(ts int64) *Measurement {
	m.l.RLock()
	defer m.l.RUnlock()

	if _, ok := m.m[ts]; !ok {
		return nil
	}
	ret := *m.m[ts]
	return &ret
}
