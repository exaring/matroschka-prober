package config

var (
	dfltBasePort = uint16(32768)
	dfltClass    = Class{
		Name: "BE",
		TOS:  0x00,
	}
	dfltTimeoutMS                     = uint64(500)
	dfltFrontend                      = ":9517"
	dfltMeasurementLengthMS           = uint64(1000)
	dfltMeasurementLengthAggregatedMS = uint64(30000)
	dfltPayloadSizeBytes              = uint64(0)
	dfltPPS                           = uint64(25)
	dfltSrcRange                      = "169.254.0.0/16"
	dfltSpoofReplySrc                 = true
)

// Config represents the configuration of matroschka-prober
type Config struct {
	Frontend *string   `yaml:"frontend"`
	Defaults *Defaults `yaml:"defaults"`
	Classes  []Class   `yaml:"classes"`
	Paths    []Path    `yaml:"paths"`
	Routers  []Router  `yaml:"routers"`
}

// Defaults represents the default section of the config
type Defaults struct {
	BasePort                      *uint16 `yaml:"base_port"`
	MeasurementLengthMS           *uint64 `yaml:"measurement_length_ms"`
	MeasurementLengthAggregatedMS *uint64 `yaml:"measurement_length_aggregated_ms"`
	PayloadSizeBytes              *uint64 `yaml:"payload_size_bytes"`
	PPS                           *uint64 `yaml:"pps"`
	SpoofReplySrc                 *bool   `yaml:"spoof_replay_src"`
	SrcRange                      *string `yaml:"src_range"`
	TimeoutMS                     *uint64 `yaml:"timeout"`
}

// Class reperesnets a traffic class in the config file
type Class struct {
	Name string `yaml:"name"`
	TOS  uint8  `yaml:"tos"`
}

// Path represents a path to be probed
type Path struct {
	Name                          string   `yaml:"name"`
	Hops                          []string `yaml:"hops"`
	BasePort                      *uint16  `yaml:"base_port"`
	MeasurementLengthMS           *uint64  `yaml:"measurement_length_ms"`
	MeasurementLengthAggregatedMS *uint64  `yaml:"measurement_length_aggregated_ms"`
	PayloadSizeBytes              *uint64  `yaml:"payload_size_bytes"`
	PPS                           *uint64  `yaml:"pps"`
	SpoofReplySrc                 *bool    `yaml:"spoof_replay_src"`
	SrcRange                      *string  `yaml:"src_range"`
	TimeoutMS                     *uint64  `yaml:"timeout"`
}

// Router represents a router used a an explicit hop in a path
type Router struct {
	Name     string `yaml:"name"`
	DstRange string `yaml:"dst_range"`
}

func (c *Config) applyDefaults() {
	if c.Defaults == nil {
		c.Defaults = &Defaults{}
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
	if p.BasePort == nil {
		p.BasePort = d.BasePort
	}

	if p.MeasurementLengthMS == nil {
		p.MeasurementLengthMS = d.MeasurementLengthMS
	}

	if p.MeasurementLengthAggregatedMS == nil {
		p.MeasurementLengthAggregatedMS = d.MeasurementLengthAggregatedMS
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
	if d.BasePort == nil {
		d.BasePort = &dfltBasePort
	}

	if d.MeasurementLengthMS == nil {
		d.MeasurementLengthMS = &dfltMeasurementLengthMS
	}

	if d.MeasurementLengthAggregatedMS == nil {
		d.MeasurementLengthAggregatedMS = &dfltMeasurementLengthAggregatedMS
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
