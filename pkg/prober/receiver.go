package prober

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/common/log"
)

func (p *Prober) receiver() {
	defer p.udpConn.Close()

	recvBuffer := make([]byte, p.mtu)
	for {
		select {
		case <-p.stop:
			return
		default:
		}

		_, err := p.udpConn.Read(recvBuffer)
		now := time.Now().UnixNano()
		if err != nil {
			log.Errorf("Unable to read from UDP socket: %v", err)
			return
		}

		atomic.AddUint64(&p.probesReceived, 1)

		pkt, err := unmarshal(recvBuffer)
		if err != nil {
			log.Errorf("Unable to unmarshal message: %v", err)
			return
		}

		err = p.transitProbes.remove(pkt.Seq)
		if err != nil {
			// Probe was count as lost, so we ignore it from here on
			continue
		}

		rtt := now - pkt.Ts
		if p.timedOut(rtt) {
			// Probe arrived late. rttTimoutChecker() will clean up after it. So we ignore it from here on
			continue
		}

		p.measurements.AddRecv(pkt.Ts, uint64(rtt), *p.path.MeasurementLengthMS)
	}
}

func (p *Prober) timedOut(s int64) bool {
	return s > int64(msToNS(*p.path.TimeoutMS))
}

func msToNS(s uint64) uint64 {
	return s * 1000000
}
