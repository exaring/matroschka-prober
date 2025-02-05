package config

import (
	"fmt"
	"math/big"
	"net"
	"net/netip"

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
	dfltListenAddress       = "0.0.0.0:9517"
	dfltMeasurementLengthMS = uint64(1000)
	dfltPayloadSizeBytes    = uint64(0)
	dfltPPS                 = uint64(25)
	dfltSrcRange            = "169.254.0.0/16"
	dflIPv6SrcRange         = "fc00::/112"
	dfltMetricsPath         = "/metrics"
)

// Config represents the configuration of matroschka-prober
type Config struct {
	// docgen:nodoc
	// this member is not configured on the yaml file
	Version string
	// description: |
	//   Path used to expose the metrics.
	MetricsPath *string `yaml:"metrcis_path,omitempty"`
	// description: |
	//   Socket to use for exposing metrics. Takes a string with the format <ip_address>:<port>.
	//   For IPv6, the string must have the format [<address>]:port.
	ListenAddressStr *string `yaml:"listen_address,omitempty"`
	// docgen:nodoc
	ListenAddress netip.AddrPort
	// description: |
	//   Base port used to listen for returned packets. If multiple paths are defined, each will take the next available port starting from <base_port>.
	BasePort *uint16 `yaml:"base_port,omitempty"`
	// description: |
	//   Default configuration parameters.
	Defaults *Defaults `yaml:"defaults,omitempty"`
	// description: |
	//   Range of IP addresses used as a source for the package. Useful to add some variance in the parameters used to hash the packets in ECMP scenarios
	//   The maximum allowed range is 2^16 addresses (/16 mask in IPv4 and /112 mask in IPv6)
	//   For IPv6, all ip addresses specified here *must* be also configured in the system.
	SrcRangeStr *string `yaml:"src_range,omitempty"`
	// docgen:nodoc
	SrcRange *net.IPNet
	// description: |
	//   Class of services.
	Classes []Class `yaml:"classes,omitempty"`
	// description: |
	//   List of paths to probe.
	Paths []Path `yaml:"paths,omitempty"`
	// description: |
	//   List of routers used as explicit hops in the path.
	Routers []Router `yaml:"routers,omitempty"`
}

// Defaults represents the default section of the config
type Defaults struct {
	// description: |
	//   Measurement interval expressed in milliseconds.
	//   IMPORTANT: If you are scraping the exposed metrics from /metrics, your scraping tool needs to scrape at least once in your defined interval.
	//   E.G if you define a measurement length of 1000ms, your scraping tool muss scrape at least 1/s, otherwise the data will be gone.
	MeasurementLengthMS *uint64 `yaml:"measurement_length_ms,omitempty"`
	// description: |
	//   Optional size of the payload (default = 0).
	PayloadSizeBytes *uint64 `yaml:"payload_size_bytes,omitempty"`
	// description: |
	//   Amount of probing packets that will be sent per second.
	PPS *uint64 `yaml:"pps,omitempty"`
	// description: |
	//   Range of IP addresses used as a source for the package. Useful to add some variance in the parameters used to hash the packets in ECMP scenarios
	//   Defaults to 169.254.0.0/16 for IPv4 and fc00::/112 for IPv6
	//   The maximum allowed range is 2^16 addresses (/16 mask in IPv4 and /112 mask in IPv6)
	//   For IPv6, all ip addresses specified here *must* be also configured in the system.
	//   If you are defining multiple paths, some which use IPv4 and some with IPv6, you must define the src_range for each router separately
	SrcRangeStr *string `yaml:"src_range,omitempty"`
	// docgen:nodoc
	SrcRange *net.IPNet
	// description: |
	//   Timeouts expressed in milliseconds
	TimeoutMS *uint64 `yaml:"timeout,omitempty"`
	// description: |
	//  Source Interface
	SrcInterface *string `yaml:"src_interface,omitempty"`
}

// Class reperesnets a traffic class in the config file
type Class struct {
	// description: |
	//   Name of the traffic class.
	Name string `yaml:"name,omitempty"`
	// description: |
	//    Type of Service assigned to the class.
	TOS uint8 `yaml:"tos,omitempty"`
}

// Path represents a path to be probed
type Path struct {
	// description: |
	//   Name for the path.
	Name string `yaml:"name,omitempty"`
	// description: |
	//   List of hops to probe.
	Hops []string `yaml:"hops,omitempty"`
	// description: |
	//   Measurement interval expressed in milliseconds.
	MeasurementLengthMS *uint64 `yaml:"measurement_length_ms,omitempty"`
	// description: |
	//   Payload size expressed in Bytes.
	PayloadSizeBytes *uint64 `yaml:"payload_size_bytes,omitempty"`
	// description: |
	//   Amount of probing packets that will be sent per second.
	PPS *uint64 `yaml:"pps,omitempty"`
	// description: |
	//   Timeout expressed in milliseconds.
	TimeoutMS *uint64 `yaml:"timeout,omitempty"`
}

// Router represents a router used a an explicit hop in a path
type Router struct {
	// description: |
	//   Name of the router.
	Name string `yaml:"name,omitempty"`
	// description: |
	//   Destination range of IP addresses.
	// Note: for IPv6 addresses, the maximum allowed range is /112
	DstRangeStr string `yaml:"dst_range,omitempty"`
	// docgen:nodoc
	DstRange *net.IPNet
	// description: |
	//   Range of source ip addresses.
	// Note: for IPv6 addresses, the maximum allowed range is /112
	SrcRangeStr string `yaml:"src_range,omitempty"`
	// docgen:nodoc
	SrcRange *net.IPNet
}

// Validate validates a configuration
func (c *Config) Validate() error {
	err := c.validatePaths()
	if err != nil {
		return fmt.Errorf("Path validation failed: %v", err)
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

// ApplyDefaults applies default settings if they are missing from loaded config.
func (c *Config) ApplyDefaults() error {
	if c.Defaults == nil {
		c.Defaults = &Defaults{}
	}
	err := c.Defaults.applyDefaults()
	if err != nil {
		return fmt.Errorf("there was a problem applying the defaults: %w", err)
	}

	if c.SrcRange == nil {
		c.SrcRangeStr = c.Defaults.SrcRangeStr
		c.SrcRange = c.Defaults.SrcRange
	}

	if c.MetricsPath == nil {
		c.MetricsPath = &dfltMetricsPath
	}

	if c.ListenAddressStr == nil {
		c.ListenAddressStr = &dfltListenAddress
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

	return nil
}

func (r *Router) applyDefaults(d *Defaults) {
	if r.SrcRangeStr == "" {
		r.SrcRangeStr = *d.SrcRangeStr
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

func (d *Defaults) applyDefaults() error {
	if d.MeasurementLengthMS == nil {
		d.MeasurementLengthMS = &dfltMeasurementLengthMS
	}

	if d.PayloadSizeBytes == nil {
		d.PayloadSizeBytes = &dfltPayloadSizeBytes
	}

	if d.PPS == nil {
		d.PPS = &dfltPPS
	}

	if d.SrcRangeStr == nil {
		d.SrcRangeStr = &dfltSrcRange
		var err error
		d.SrcRange, err = convertIPRange(*d.SrcRangeStr)
		if err != nil {
			return fmt.Errorf("there was an erro parsing src_range: %w", err)
		}
	}

	if d.TimeoutMS == nil {
		d.TimeoutMS = &dfltTimeoutMS
	}

	return nil
}

// GetConfiguredSrcAddr gets an IPv4 address of the configured src interface
func (c *Config) GetConfiguredSrcAddr() (net.IP, error) {
	if c.Defaults.SrcInterface == nil {
		return nil, nil
	}

	ipVersion := GetIPVersion(c.SrcRange)

	return GetInterfaceAddr(*c.Defaults.SrcInterface, ipVersion)
}

func GetIPVersion(network *net.IPNet) uint8 {
	version := network.IP.To4()
	if version != nil {
		return 4
	}

	version = network.IP.To16()
	if version != nil {
		return 6
	}

	panic(fmt.Sprintf("Couldn't determine the protocol version for address %s", version))
}

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
func (c *Config) PathToProberHops(pathCfg Path) ([]prober.Hop, error) {
	res := make([]prober.Hop, 0)

	for _, hop := range pathCfg.Hops {
		r := getRouter(c.Routers, hop)
		if r == nil {
			return nil, fmt.Errorf("unable to find hop %s", hop)
		}

		h := prober.Hop{
			Name:     r.Name,
			DstRange: GenerateAddrs(r.DstRange),
			SrcRange: GenerateAddrs(r.SrcRange),
		}
		res = append(res, h)

	}

	return res, nil
}

func getRouter(haystack []Router, name string) *Router {
	for _, r := range haystack {
		if r.Name == name {
			return &r
		}
	}

	return nil
}

// GenerateAddrs returns a list of all IPs in addrRange
func GenerateAddrs(addrRange *net.IPNet) []net.IP {
	ret, err := generateIPList(addrRange)
	if err != nil {
		panic(err)
	}

	return ret
}

// calculateSubnetSize calculates the number of IP addresses in a subnet
func calculateSubnetSize(subnet *net.IPNet) (uint32, error) {
	ones, bits := subnet.Mask.Size()
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
func generateIPList(network *net.IPNet) ([]net.IP, error) {
	var ipList []net.IP
	version := int(GetIPVersion(network))

	ip := big.NewInt(0)
	ip.SetBytes(network.IP)

	// Calculate the number of IP addresses in the network
	ones, bits := network.Mask.Size()
	numIPs := big.NewInt(0)
	numIPs.Exp(big.NewInt(2), big.NewInt(int64(bits-ones)), nil)

	// Iterate over all IP addresses in the network and add them to the list
	for i := big.NewInt(0); i.Cmp(numIPs) < 0; i.Add(i, big.NewInt(1)) {
		// Convert the big.Int back to an IP address
		ipBytes := ip.Bytes()
		if len(ipBytes) < version {
			// Pad the IP address with zeros if necessary
			ipBytes = append(make([]byte, version-len(ipBytes)), ipBytes...)
		}
		ipList = append(ipList, net.IP(ipBytes))

		// Increment the IP address
		ip.Add(ip, big.NewInt(1))
	}

	return ipList, nil
}

func (c *Config) ConvertIPAddresses() error {
	var err error
	if c.ListenAddressStr != nil {
		c.ListenAddress, err = stringToAddrPort(*c.ListenAddressStr)
		if err != nil {
			return fmt.Errorf("there was an error parsing listen_addres: %w", err)
		}
	}

	if c.SrcRangeStr != nil {
		c.SrcRange, err = convertIPRange(*c.SrcRangeStr)
		if err != nil {
			return fmt.Errorf("there was an erro parsing src_range: %w", err)
		}
	}

	c.Defaults.SrcRange, err = convertIPRange(*c.Defaults.SrcRangeStr)
	if err != nil {
		return fmt.Errorf("there was an error parsing defaults.src_range: %w", err)
	}

	for key, router := range c.Routers {
		c.Routers[key].DstRange, err = convertIPRange(router.DstRangeStr)
		if err != nil {
			return fmt.Errorf("there was an error parsing routers.dst_range: %w", err)
		}

		c.Routers[key].SrcRange, err = convertIPRange(router.SrcRangeStr)
		if err != nil {
			return fmt.Errorf("there was an error parsing router.src_range: %w", err)
		}
	}

	return nil
}

func convertIPAddress(s string) (net.IP, error) {
	ip, _, err := net.ParseCIDR(s)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return ip, nil
}

func stringToAddrPort(s string) (netip.AddrPort, error) {
	addr, err := netip.ParseAddrPort(s)
	if err != nil {
		return addr, fmt.Errorf("%w", err)
	}

	return addr, nil
}

func convertIPRange(s string) (*net.IPNet, error) {
	_, ipRange, err := net.ParseCIDR(s)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return ipRange, nil
}

func initDefaultRange(ip string) net.IPNet {
	_, ipRange, err := net.ParseCIDR(ip)
	if err != nil {
		panic("error parsing default source range in the code")
	}

	return *ipRange
}
