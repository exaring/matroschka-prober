package probermanager

import (
	"fmt"
	"sync"

	"github.com/exaring/matroschka-prober/pkg/config"
	"github.com/exaring/matroschka-prober/pkg/prober"
	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
)

type proberIndex struct {
	path string
	tos  uint8
}

type ProberManager struct {
	probers   map[proberIndex]*prober.Prober
	probersMu sync.RWMutex
}

func New() *ProberManager {
	return &ProberManager{
		probers: make(map[proberIndex]*prober.Prober),
	}
}

func (pm *ProberManager) Configure(cfg *config.Config) error {
	pm.probersMu.Lock()
	defer pm.probersMu.Unlock()

	desiredProbers, err := getDesiredProbers(cfg)
	if err != nil {
		return fmt.Errorf("unable to create probers: %v", err)
	}

	pm.adjustProbers(desiredProbers)
	return nil
}

func getDesiredProbers(cfg *config.Config) (map[proberIndex]*prober.Prober, error) {
	confSrc, err := cfg.GetConfiguredSrcAddr()
	if err != nil {
		return nil, fmt.Errorf("Unable to get configured src addr: %v", err)
	}

	desiredProbers := make(map[proberIndex]*prober.Prober, 0)
	for i := range cfg.Paths {
		for j := range cfg.Classes {
			hops, err := cfg.PathToProberHops(cfg.Paths[i])
			if err != nil {
				log.Errorf("Unable to create hops %v for path %q: %v", cfg.Paths[i].Hops, cfg.Paths[i].Name, err)
				continue
			}

			pcfg := prober.Config{
				Name:              cfg.Paths[i].Name,
				BasePort:          *cfg.BasePort + uint16(i),
				ConfiguredSrcAddr: confSrc,
				SrcAddrs:          config.GenerateAddrs(cfg.SrcRange),
				Hops:              hops,
				StaticLabels:      labels(cfg.Paths[i].Labels),
				TOS: prober.TOS{
					Name:  cfg.Classes[j].Name,
					Value: cfg.Classes[j].TOS,
				},
				PPS:                 *cfg.Paths[i].PPS,
				PayloadSizeBytes:    *cfg.Paths[i].PayloadSizeBytes,
				MeasurementLengthMS: *cfg.Paths[i].MeasurementLengthMS,
				TimeoutMS:           *cfg.Paths[i].TimeoutMS,
				IPVersion:           config.GetIPVersion(cfg.SrcRange),
			}

			desiredProbers[proberIndex{
				path: pcfg.Name,
				tos:  pcfg.TOS.Value,
			}] = prober.New(pcfg)

		}
	}

	return desiredProbers, nil
}

func (pm *ProberManager) adjustProbers(desiredProbers map[proberIndex]*prober.Prober) {
	pm.startMissingProbers(desiredProbers)
	pm.stopAbandonedProbers(desiredProbers)
	pm.recreateChangedProbers(desiredProbers)
}

func (pm *ProberManager) startMissingProbers(desiredProbers map[proberIndex]*prober.Prober) {
	for k, p := range desiredProbers {
		if _, exists := pm.probers[k]; !exists {
			log.Infof("Adding prober %q (0x%x)", k.path, k.tos)
			pm.probers[k] = p
			err := p.Start()
			if err != nil {
				log.Errorf("unable to start prober %s (0x%x)", k.path, k.tos)
			}
		}
	}
}

func (pm *ProberManager) stopAbandonedProbers(desiredProbers map[proberIndex]*prober.Prober) {
	for k, p := range pm.probers {
		if _, isDesired := desiredProbers[k]; !isDesired {
			log.Infof("Removing prober %q (0x%x)", k.path, k.tos)
			p.Stop()
			delete(pm.probers, k)
		}
	}
}

func (pm *ProberManager) recreateChangedProbers(desiredProbers map[proberIndex]*prober.Prober) {
	for k, p := range desiredProbers {
		pm.probers[k].Config()
		if pm.probers[k].Config().Equal(desiredProbers[k].Config()) {
			continue
		}

		log.Infof("Reconfiguring prober %q (0x%x)", k.path, k.tos)
		pm.probers[k].Stop()
		pm.probers[k] = p
		err := pm.probers[k].Start()
		if err != nil {
			log.Errorf("unable to start prober %q (0x%x): %v", k.path, k.tos, err)
		}
	}
}

func (pm *ProberManager) GetCollectors() []prometheus.Collector {
	pm.probersMu.RLock()
	defer pm.probersMu.RUnlock()

	ret := make([]prometheus.Collector, 0, len(pm.probers))
	for i := range pm.probers {
		ret = append(ret, pm.probers[i])
	}

	return ret
}

func labels(m map[string]string) []prober.Label {
	ret := make([]prober.Label, 0)
	for k, v := range m {
		ret = append(ret, prober.Label{
			Key:   k,
			Value: v,
		})
	}

	return ret
}
