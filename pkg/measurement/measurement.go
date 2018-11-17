package measurement

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Measurement represents a measurement
type Measurement struct {
	Sent     uint64
	Received uint64
	RTTSum   uint64
	RTTMin   uint64
	RTTMax   uint64
	RTTs     []uint64
}

// MeasurementsDB manages measurements
type MeasurementsDB struct {
	m map[int64]*Measurement
	l sync.RWMutex
}

// NewDB creates a new measurements database
func NewDB() *MeasurementsDB {
	return &MeasurementsDB{
		m: make(map[int64]*Measurement),
	}
}

// AddSent adds a sent probe to the db
func (m *MeasurementsDB) AddSent(ts int64) {
	m.l.Lock()

	if m.m[ts] == nil {
		m.m[ts] = &Measurement{
			RTTs: make([]uint64, 0),
		}
	}
	m.m[ts].Sent++

	m.l.Unlock() // This is not defered for performance reason
}

// AddRecv adds a received probe to the db
func (m *MeasurementsDB) AddRecv(sentTs int64, rtt uint64, deltaT uint64) {
	m.l.RLock()

	allignedTs := sentTs - sentTs%int64(deltaT)
	if _, ok := m.m[allignedTs]; !ok {
		log.Debugf("Received probe at %d sent at %d with rtt %d after bucket %d was removed. Now=%d", sentTs+int64(rtt), sentTs, allignedTs, rtt, time.Now().UnixNano())
		m.l.RUnlock() // This is not defered for performance reason
		return
	}

	m.m[allignedTs].Received++
	m.m[allignedTs].RTTs = append(m.m[allignedTs].RTTs, rtt)
	m.m[allignedTs].RTTSum += rtt

	if rtt < m.m[allignedTs].RTTMin || m.m[allignedTs].RTTMin == 0 {
		m.m[allignedTs].RTTMin = rtt
	}

	if rtt > m.m[allignedTs].RTTMax {
		m.m[allignedTs].RTTMax = rtt
	}

	m.l.RUnlock() // This is not defered for performance reason
}

// RemoveOlder removes all probes from the db that are older than ts
func (m *MeasurementsDB) RemoveOlder(ts int64) {
	m.l.Lock()
	defer m.l.Unlock()

	for t := range m.m {
		if t < ts {
			delete(m.m, t)
		}
	}
}

// Get get's the measurement at ts
func (m *MeasurementsDB) Get(ts int64) *Measurement {
	m.l.RLock()
	defer m.l.RUnlock()

	if _, ok := m.m[ts]; !ok {
		return nil
	}

	ret := *m.m[ts]
	return &ret
}
