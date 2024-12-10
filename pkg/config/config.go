package config

import (
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
	// docgen:nodoc
	// this member is not configured on the yaml file
	Version string
	// description: |
	//   Path used to expose the metrics.
	MetricsPath *string `yaml:"metrcis_path"`
	// description: |
	//   Socket to use for exposing metrics. Takes a string with the format <ip_address>:<port>.
	//   For IPv6, the string must have the format [<address>]:port.
	ListenAddress *string `yaml:"listen_address"`
	// description: |
	//   Base port used to listen for returned packets. If multiple paths are defined, each will take the next available port starting from <base_port>.
	BasePort *uint16 `yaml:"base_port"`
	// description: |
	//   Default configuration parameters.
	Defaults *Defaults `yaml:"defaults"`
	// description: |
	//   Range of IP addresses used as a source for the package. Useful to add some variance in the parameters used to hash the packets in ECMP scenarios
	//   The maximum allowed range is 2^16 addresses (/16 mask in IPv4 and /112 mask in IPv6)
	//   For IPv6, all ip addresses specified here *must* be also configured in the system.
	SrcRange *string `yaml:"src_range"`
	// description: |
	//   Class of services.
	Classes []Class `yaml:"classes"`
	// description: |
	//   List of paths to probe.
	Paths []Path `yaml:"paths"`
	// description: |
	//   List of routers used as explicit hops in the path.
	Routers []Router `yaml:"routers"`
}

// Defaults represents the default section of the config
type Defaults struct {
	// description: |
	//   Measurement interval expressed in milliseconds.
	//   IMPORTANT: If you are scraping the exposed metrics from /metrics, your scraping tool needs to scrape at least once in your defined interval.
	//   E.G if you define a measurement length of 1000ms, your scraping tool muss scrape at least 1/s, otherwise the data will be gone.
	MeasurementLengthMS *uint64 `yaml:"measurement_length_ms"`
	// description: |
	//   Optional size of the payload (default = 0).
	PayloadSizeBytes *uint64 `yaml:"payload_size_bytes"`
	// description: |
	//   Amount of probing packets that will be sent per second.
	PPS *uint64 `yaml:"pps"`
	// description: |
	//   Range of IP addresses used as a source for the package. Useful to add some variance in the parameters used to hash the packets in ECMP scenarios
	//   Defaults to 169.254.0.0/16 for IPv4 and fe80::/112 for IPv6
	//   The maximum allowed range is 2^16 addresses (/16 mask in IPv4 and /112 mask in IPv6)
	//   For IPv6, all ip addresses specified here *must* be also configured in the system.
	SrcRange *string `yaml:"src_range"`
	// description: |
	//   Timeouts expressed in milliseconds
	TimeoutMS *uint64 `yaml:"timeout"`
	// description: |
	//  Source Interface
	SrcInterface *string `yaml:"src_interface"`
}

// Class reperesnets a traffic class in the config file
type Class struct {
	// description: |
	//   Name of the traffic class.
	Name string `yaml:"name"`
	// description: |
	//    Type of Service assigned to the class.
	TOS uint8 `yaml:"tos"`
}

// Path represents a path to be probed
type Path struct {
	// description: |
	//   Name for the path.
	Name string `yaml:"name"`
	// description: |
	//   List of hops to probe.
	Hops []string `yaml:"hops"`
	// description: |
	//   Measurement interval expressed in milliseconds.
	MeasurementLengthMS *uint64 `yaml:"measurement_length_ms"`
	// description: |
	//   Payload size expressed in Bytes.
	PayloadSizeBytes *uint64 `yaml:"payload_size_bytes"`
	// description: |
	//   Amount of probing packets that will be sent per second.
	PPS *uint64 `yaml:"pps"`
	// description: |
	//   Timeout expressed in milliseconds.
	TimeoutMS *uint64 `yaml:"timeout"`
}

// Router represents a router used a an explicit hop in a path
type Router struct {
	// description: |
	//   Name of the router.
	Name string `yaml:"name"`
	// description: |
	//   Destination range of IP addresses.
	// Note: for IPv6 addresses, the maximum allowed range is /64
	DstRange string `yaml:"dst_range"`
	// description: |
	//   Range of source ip addresses.
	// Note: for IPv6 addresses, the maximum allowed range is /64
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

	ipVersion := c.GetIPVersion()

	return GetInterfaceAddr(*c.Defaults.SrcInterface, ipVersion)
}

func (c *Config) GetIPVersion() uint8 {
	strValue := *c.SrcRange
	ip ,_ , err := net.ParseCIDR(strValue)
	if ip == nil {
		ip = net.IP(*c.Defaults.SrcRange)
	}

	if err != nil {
		panic("No source range defined")
	}

	version := ip.To4()
	if version != nil {
		return 4
	}

	version = ip.To16()
	if version != nil {
		return 6
	}

	panic(fmt.Sprintf("Couldn't determine the protocol version for address %s", ip))
}


// TODO: return ipv6 ip for v6 prober
// GetInterfaceAddr gets an interface first IPv4 address
func GetInterfaceAddr(ifName string, ipVersion uint8) (net.IP, error) {
	ifa, err := net.InterfaceByName(ifName)
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

		if (ipVersion == 4 && ip.To4() != nil) || (ipVersion == 6 && ip.To4() == nil) {
			return ip, nil
		}
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
	maskLength, err := calculateSubnetSize(addrRange)
	if err != nil {
		panic(err)
	}

	ret, err := generateIPList(addrRange, maskLength)
	if err != nil {
		panic(err)
	}

	return ret
}

// calculateSubnetSize calculates the number of IP addresses in a subnet
func calculateSubnetSize(subnet string) (uint32, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return 0, fmt.Errorf("invalid subnet: %v", err)
	}

	ones, bits := ipNet.Mask.Size()
	if ones == 0 && bits == 0 {
		return 0, fmt.Errorf("invalid subnet mask")
	}

	// Calculate the number of IP addresses in the subnet
	numIPs := uint32(1) << uint(bits-ones)

	// Check if the number of IPs exceeds 2^16
	if numIPs > (1 << 16) {
		return 0, fmt.Errorf("number of IP addresses exceeds 2^16")
	}

	return numIPs, nil
}

// incrementIP increments an IP address by one
func incrementIP(ip net.IP) net.IP {
	incIP := make(net.IP, len(ip))
	copy(incIP, ip)
	for j := len(incIP) - 1; j >= 0; j-- {
		incIP[j]++
		if incIP[j] != 0 {
			break
		}
	}
	return incIP
}

// generateIPList generates a list of IP addresses starting from baseIP
func generateIPList(baseIP string, x uint32) ([]net.IP, error) {
	if x > (1 << 16) {
		return nil, fmt.Errorf("number of IP addresses exceeds 2^16")
	}

	ip, network, err := net.ParseCIDR(baseIP)
	ip = ip.Mask(network.Mask)
	if err != nil {
		return nil, fmt.Errorf("invalid base IP address")
	}

	ipList := make([]net.IP, x)
	for i := uint32(0); i < x; i++ {
		ipList[i] = ip
		ip = incrementIP(ip)
	}

	return ipList, nil
}
