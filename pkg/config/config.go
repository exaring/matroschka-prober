package config

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/exaring/matroschka-prober/pkg/prober"
	"github.com/pkg/errors"
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
	dfltMetricsPath         = "/metrics"
)

// Config represents the configuration of matroschka-prober
type Config struct {
	Version       string
	MetricsPath   *string   `yaml:"metrcis_path"`
	ListenAddress *string   `yaml:"listen_address"`
	BasePort      *uint16   `yaml:"base_port"`
	Defaults      *Defaults `yaml:"defaults"`
	SrcRange      *string   `yaml:"src_range"`
	Classes       []Class   `yaml:"classes"`
	Paths         []Path    `yaml:"paths"`
	Routers       []Router  `yaml:"routers"`
}

// Defaults represents the default section of the config
type Defaults struct {
	MeasurementLengthMS *uint64 `yaml:"measurement_length_ms"`
	PayloadSizeBytes    *uint64 `yaml:"payload_size_bytes"`
	PPS                 *uint64 `yaml:"pps"`
	SrcRange            *string `yaml:"src_range"`
	TimeoutMS           *uint64 `yaml:"timeout"`
	SrcInterface        *string `yaml:"src_interface"`
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
	TimeoutMS           *uint64  `yaml:"timeout"`
}

// Router represents a router used a an explicit hop in a path
type Router struct {
	Name     string `yaml:"name"`
	DstRange string `yaml:"dst_range"`
	SrcRange string `yaml:"src_range"`
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
	c.Defaults.applyDefaults()

	if c.SrcRange == nil {
		c.SrcRange = c.Defaults.SrcRange
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

	for i := range c.Paths {
		c.Paths[i].applyDefaults(c.Defaults)
	}

	for i := range c.Routers {
		c.Routers[i].applyDefaults(c.Defaults)
	}

	if c.Classes == nil {
		c.Classes = []Class{
			dfltClass,
		}
	}
}

func (r *Router) applyDefaults(d *Defaults) {
	if r.SrcRange == "" {
		r.SrcRange = *d.SrcRange
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

	if d.SrcRange == nil {
		d.SrcRange = &dfltSrcRange
	}

	if d.TimeoutMS == nil {
		d.TimeoutMS = &dfltTimeoutMS
	}
}

// GetConfiguredSrcAddr gets an IPv4 address of the configured src interface
func (c *Config) GetConfiguredSrcAddr() (net.IP, error) {
	if c.Defaults.SrcInterface == nil {
		return nil, nil
	}

	ifa, err := net.InterfaceByName(*c.Defaults.SrcInterface)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get interface")
	}

	addrs, err := ifa.Addrs()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get addresses")
	}

	for _, a := range addrs {
		ip, _, err := net.ParseCIDR(a.String())
		if err != nil {
			continue
		}

		if ip.To4() == nil {
			continue
		}

		return ip, nil
	}

	return nil, nil
}

// PathToProberHops generates prober hops
func (c *Config) PathToProberHops(pathCfg Path) []prober.Hop {
	res := make([]prober.Hop, 0)

	for i := range pathCfg.Hops {
		for j := range c.Routers {
			if pathCfg.Hops[i] != c.Routers[j].Name {
				continue
			}

			h := prober.Hop{
				Name:     c.Routers[j].Name,
				DstRange: GenerateAddrs(c.Routers[j].DstRange),
				SrcRange: GenerateAddrs(c.Routers[j].SrcRange),
			}
			res = append(res, h)
		}
	}

	return res
}

// GenerateAddrs returns a list of all IPs in addrRange
func GenerateAddrs(addrRange string) []net.IP {
	_, n, err := net.ParseCIDR(addrRange)
	if err != nil {
		panic(err)
	}

	baseAddr := getCIDRBase(*n)
	c := maskAddrCount(*n)
	ret := make([]net.IP, c)

	for i := uint32(0); i < c; i++ {
		ret[i] = net.IP(uint32Byte(baseAddr + i%c))
	}

	return ret
}

func getCIDRBase(n net.IPNet) uint32 {
	return uint32b(n.IP)
}

func uint32b(data []byte) (ret uint32) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &ret)
	return
}

func getNthAddr(n net.IPNet, i uint32) net.IP {
	baseAddr := getCIDRBase(n)
	c := maskAddrCount(n)
	return net.IP(uint32Byte(baseAddr + i%c))
}

func maskAddrCount(n net.IPNet) uint32 {
	ones, bits := n.Mask.Size()
	if ones == bits {
		return 1
	}

	x := uint32(1)
	for i := ones; i < bits; i++ {
		x = x * 2
	}
	return x
}

func uint32Byte(data uint32) (ret []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, data)
	return buf.Bytes()
}
