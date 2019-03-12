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

// AddSentAndRemoveOlder adds a sent probe to the db and removes all probes from the db that are older than removeOlderTs
func (m *MeasurementsDB) AddSentAndRemoveOlder(ts int64, removeOlderTs int64) {
	m.l.Lock()

	if m.m[ts] == nil {
		m.m[ts] = &Measurement{
			RTTs: make([]uint64, 0),
		}
		if (removeOlderTs > 0) {
			for t := range m.m {
				if t < removeOlderTs {
					delete(m.m, t)
				}
			}
		}
	}
	m.m[ts].Sent++

	m.l.Unlock() // This is not defered for performance reason
}

// AddRecv adds a received probe to the db
func (m *MeasurementsDB) AddRecv(sentTsNS int64, rtt uint64, measurementDurationMS uint64) {
	m.l.RLock()

	allignedTs := sentTsNS - sentTsNS%int64(measurementDurationMS*uint64(time.Millisecond))
	if _, ok := m.m[allignedTs]; !ok {
		log.Debugf("Received probe at %d sent at %d with rtt %d after bucket %d was removed. Now=%d", sentTsNS+int64(rtt), sentTsNS, allignedTs, rtt, time.Now().UnixNano())
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
