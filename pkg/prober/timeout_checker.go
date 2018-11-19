package prober

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func (p *Prober) rttTimeoutChecker() {
	t := time.NewTicker(time.Duration(*p.path.MeasurementLengthMS) * time.Millisecond)

	for {
		<-t.C

		timeout := *p.path.MeasurementLengthMS * uint64(time.Millisecond)
		maxTS := (uint64(time.Now().UnixNano()) - 3*timeout)
		for s := range p.transitProbes.getLt(int64(maxTS)) {
			err := p.transitProbes.remove(s)
			if err != nil {
				log.Infof("Probe %d timeouted: Unable to remove: %v", s, err)
			}
		}
	}
}
