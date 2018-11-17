package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/exaring/matroschka-prober/pkg/config"
	"github.com/exaring/matroschka-prober/pkg/prober"
	log "github.com/sirupsen/logrus"
)

var (
	cfgFilepath = flag.String("config.file", "matroschka.yml", "Config file")
)

func main() {
	flag.Parse()

	cfg, err := loadConfig(*cfgFilepath)
	if err != nil {
		log.Errorf("Unable to load config: %v", err)
		os.Exit(1)
	}

	for i := range cfg.Paths {
		for j := range cfg.Classes {
			p := prober.New(cfg, cfg.Paths[i], cfg.Classes[j].TOS)
			p.Start()
		}

	}
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
