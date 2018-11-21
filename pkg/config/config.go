package config

import (
	"fmt"
	"net"
)

var (
	dfltBasePort = uint16(32768)
	dfltClass    = Class{
		Name: "BE",
		TOS:  0x00,
	}
	dfltTimeoutMS           = uint64(500)
	dfltListenAddress       = ":9517"
	dfltMeasurementLengthMS = uint64(1000)
	dfltPayloadSizeBytes    = uint64(0)
	dfltPPS                 = uint64(25)
	dfltSrcRange            = "169.254.0.0/16"
	dfltSpoofReplySrc       = true
	dfltMetricsPath         = "/metrics"
)

// Config represents the configuration of matroschka-prober
type Config struct {
	Version       string
	MetricsPath   *string   `yaml:"metrcis_path"`
	ListenAddress *string   `yaml:"listen_address"`
	BasePort      *uint16   `yaml:"base_port"`
	Defaults      *Defaults `yaml:"defaults"`
	Classes       []Class   `yaml:"classes"`
	Paths         []Path    `yaml:"paths"`
	Routers       []Router  `yaml:"routers"`
}

// Defaults represents the default section of the config
type Defaults struct {
	MeasurementLengthMS *uint64 `yaml:"measurement_length_ms"`
	PayloadSizeBytes    *uint64 `yaml:"payload_size_bytes"`
	PPS                 *uint64 `yaml:"pps"`
	SpoofReplySrc       *bool   `yaml:"spoof_reply_src"`
	SrcRange            *string `yaml:"src_range"`
	TimeoutMS           *uint64 `yaml:"timeout"`
}

// Class reperesnets a traffic class in the config file
type Class struct {
	Name string `yaml:"name"`
	TOS  uint8  `yaml:"tos"`
}

// Path represents a path to be probed
type Path struct {
	Name                string   `yaml:"name"`
	Hops                []string `yaml:"hops"`
	MeasurementLengthMS *uint64  `yaml:"measurement_length_ms"`
	PayloadSizeBytes    *uint64  `yaml:"payload_size_bytes"`
	PPS                 *uint64  `yaml:"pps"`
	SpoofReplySrc       *bool    `yaml:"spoof_reply_src"`
	SrcRange            *string  `yaml:"src_range"`
	TimeoutMS           *uint64  `yaml:"timeout"`
}

// Router represents a router used a an explicit hop in a path
type Router struct {
	Name     string `yaml:"name"`
	DstRange string `yaml:"dst_range"`
}

// Validate validates a configuration
func (c *Config) Validate() error {
	err := c.validatePaths()
	if err != nil {
		return fmt.Errorf("Path validation failed: %v", err)
	}

	err = c.validateRouters()
	if err != nil {
		return fmt.Errorf("Router validation failed: %v", err)
	}

	return nil
}

func (c *Config) validatePaths() error {
	for i := range c.Paths {
		for j := range c.Paths[i].Hops {
			if !c.routerExists(c.Paths[i].Hops[j]) {
				return fmt.Errorf("Router %q of path %q does not exist", c.Paths[i].Hops[j], c.Paths[i].Name)
			}
		}
	}

	return nil
}

func (c *Config) routerExists(needle string) bool {
	for i := range c.Routers {
		if c.Routers[i].Name == needle {
			return true
		}
	}

	return false
}

func (c *Config) validateRouters() error {
	for i := range c.Routers {
		_, _, err := net.ParseCIDR(c.Routers[i].DstRange)
		if err != nil {
			return fmt.Errorf("Unable to parse dst IP range for router %q: %v", c.Routers[i].Name, err)
		}
	}

	return nil
}

// ApplyDefaults applies default settings if they are missing from loaded config.
func (c *Config) ApplyDefaults() {
	if c.Defaults == nil {
		c.Defaults = &Defaults{}
	}

	if c.MetricsPath == nil {
		c.MetricsPath = &dfltMetricsPath
	}

	if c.ListenAddress == nil {
		c.ListenAddress = &dfltListenAddress
	}

	if c.BasePort == nil {
		c.BasePort = &dfltBasePort
	}

	c.Defaults.applyDefaults()

	for i := range c.Paths {
		c.Paths[i].applyDefaults(c.Defaults)
	}

	if c.Classes == nil {
		c.Classes = []Class{
			dfltClass,
		}
	}
}

func (p *Path) applyDefaults(d *Defaults) {
	if p.MeasurementLengthMS == nil {
		p.MeasurementLengthMS = d.MeasurementLengthMS
	}

	if p.PayloadSizeBytes == nil {
		p.PayloadSizeBytes = d.PayloadSizeBytes
	}

	if p.PPS == nil {
		p.PPS = d.PPS
	}

	if p.SpoofReplySrc == nil {
		p.SpoofReplySrc = d.SpoofReplySrc
	}

	if p.SrcRange == nil {
		p.SrcRange = d.SrcRange
	}

	if p.TimeoutMS == nil {
		p.TimeoutMS = d.TimeoutMS
	}
}

func (d *Defaults) applyDefaults() {
	if d.MeasurementLengthMS == nil {
		d.MeasurementLengthMS = &dfltMeasurementLengthMS
	}

	if d.PayloadSizeBytes == nil {
		d.PayloadSizeBytes = &dfltPayloadSizeBytes
	}

	if d.PPS == nil {
		d.PPS = &dfltPPS
	}

	if d.SpoofReplySrc == nil {
		d.SpoofReplySrc = &dfltSpoofReplySrc
	}

	if d.SrcRange == nil {
		d.SrcRange = &dfltSrcRange
	}

	if d.TimeoutMS == nil {
		d.TimeoutMS = &dfltTimeoutMS
	}
}
