package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/exaring/matroschka-prober/pkg/config"
	"github.com/exaring/matroschka-prober/pkg/frontend"
	"github.com/exaring/matroschka-prober/pkg/prober"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	_ "net/http/pprof"
)

var (
	cfgFilepath = flag.String("config.file", "matroschka.yml", "Config file")
	logLevel    = flag.String("log.level", "debug", "Log Level")
)

func main() {
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Errorf("Unable to parse log.level: %v", err)
		os.Exit(1)
	}
	log.SetLevel(level)

	cfg, err := loadConfig(*cfgFilepath)
	if err != nil {
		log.Errorf("Unable to load config: %v", err)
		os.Exit(1)
	}

	confSrc, err := cfg.GetConfiguredSrcAddr()
	if err != nil {
		log.Errorf("Unable to get configured src addr: %v", err)
		os.Exit(1)
	}

	probers := make([]*prober.Prober, 0)
	for i := range cfg.Paths {
		for j := range cfg.Classes {
			log.Infof("Starting prober for path %q class %q", cfg.Paths[i].Name, cfg.Classes[j].Name)
			p, err := prober.New(prober.Config{
				BasePort:          *cfg.BasePort,
				ConfiguredSrcAddr: confSrc,
				SrcAddrs:          config.GenerateAddrs(*cfg.SrcRange),
				Hops:              cfg.PathToProberHops(cfg.Paths[i]),
				StaticLabels:      []prober.Label{},
				TOS: prober.TOS{
					Name:  cfg.Classes[j].Name,
					Value: cfg.Classes[j].TOS,
				},
				PPS:                 *cfg.Paths[i].PPS,
				PayloadSizeBytes:    *cfg.Paths[i].PayloadSizeBytes,
				MeasurementLengthMS: *cfg.Paths[i].MeasurementLengthMS,
				TimeoutMS:           *cfg.Paths[i].TimeoutMS,
				IPProtocol:          cfg.GetIPVersion(),
			})

			if err != nil {
				log.Errorf("Unable to get new prober: %v", err)
				os.Exit(1)
			}

			err = p.Start()
			if err != nil {
				log.Errorf("Unable to start prober: %v", err)
				os.Exit(1)
			}
			probers = append(probers, p)
		}
	}

	fe := frontend.New(&frontend.Config{
		Version:       cfg.Version,
		MetricsPath:   *cfg.MetricsPath,
		ListenAddress: *cfg.ListenAddress,
	}, newRegistry(probers))
	go fe.Start()
	select {}
}

type registry struct {
	probers []*prober.Prober
}

func newRegistry(probers []*prober.Prober) *registry {
	return &registry{
		probers: probers,
	}
}

func (r *registry) GetCollectors() []prometheus.Collector {
	ret := make([]prometheus.Collector, len(r.probers))
	for i := range r.probers {
		ret[i] = r.probers[i]
	}

	return ret
}

func loadConfig(path string) (*config.Config, error) {
	cfgFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %q: %v", path, err)
	}

	cfg := &config.Config{}
	err = yaml.Unmarshal(cfgFile, cfg)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal: %v", err)
	}

	cfg.ApplyDefaults()
	return cfg, nil
}
