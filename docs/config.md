



## Config
Config represents the configuration of matroschka-prober






<hr />

<div class="dd">

<code>metrcis_path</code>  <i>string</i>

</div>
<div class="dt">

Path used to expose the metrics.

</div>

<hr />

<div class="dd">

<code>listen_address</code>  <i>string</i>

</div>
<div class="dt">

Address used to listen for returned packets

</div>

<hr />

<div class="dd">

<code>base_port</code>  <i>uint16</i>

</div>
<div class="dt">

Port used to listen for returned packets

</div>

<hr />

<div class="dd">

<code>defaults</code>  <i><a href="#defaults">Defaults</a></i>

</div>
<div class="dt">

Default configuration parameters

</div>

<hr />

<div class="dd">

<code>src_range</code>  <i>string</i>

</div>
<div class="dt">

Range of IP addresses used as a source for the package. Useful to add some variance in the parameters used to hash the packets in ECMP scenarios
The maximum allowed range is 2^16 addresses (/16 mask in IPv4 and /112 mask in IPv6)
For IPv6, all ip addresses specified here *must* be also configured in the system.

</div>

<hr />

<div class="dd">

<code>classes</code>  <i>[]<a href="#class">Class</a></i>

</div>
<div class="dt">

Class of services

</div>

<hr />

<div class="dd">

<code>paths</code>  <i>[]<a href="#path">Path</a></i>

</div>
<div class="dt">

List of paths to probe

</div>

<hr />

<div class="dd">

<code>routers</code>  <i>[]<a href="#router">Router</a></i>

</div>
<div class="dt">

List of routers used as explicit hops in the path.

</div>

<hr />





## Defaults
Defaults represents the default section of the config

Appears in:


- <code><a href="#config">Config</a>.defaults</code>





<hr />

<div class="dd">

<code>measurement_length_ms</code>  <i>uint64</i>

</div>
<div class="dt">

Measurement interval expressed in milliseconds.
IMPORTANT: If you are scraping the exposed metrics from /metrics, your scraping tool needs to scrape at least once in your defined interval.
E.G if you define a measurement length of 1000ms, your scraping tool muss scrape at least 1/s, otherwise the data will be gone.

</div>

<hr />

<div class="dd">

<code>payload_size_bytes</code>  <i>uint64</i>

</div>
<div class="dt">

Optional size of the payload (default = 0).

</div>

<hr />

<div class="dd">

<code>pps</code>  <i>uint64</i>

</div>
<div class="dt">

Amount of probing packets that will be sent per second.

</div>

<hr />

<div class="dd">

<code>src_range</code>  <i>string</i>

</div>
<div class="dt">

Range of IP addresses used as a source for the package. Useful to add some variance in the parameters used to hash the packets in ECMP scenarios
Defaults to 169.254.0.0/16 for IPv4 and fe80::/112 for IPv6
The maximum allowed range is 2^16 addresses (/16 mask in IPv4 and /112 mask in IPv6)
For IPv6, all ip addresses specified here *must* be also configured in the system.

</div>

<hr />

<div class="dd">

<code>timeout</code>  <i>uint64</i>

</div>
<div class="dt">

Timeouts expressed in milliseconds

</div>

<hr />

<div class="dd">

<code>src_interface</code>  <i>string</i>

</div>
<div class="dt">

Source Interface

</div>

<hr />





## Class
Class reperesnets a traffic class in the config file

Appears in:


- <code><a href="#config">Config</a>.classes</code>





<hr />

<div class="dd">

<code>name</code>  <i>string</i>

</div>
<div class="dt">

Name of the traffic class.

</div>

<hr />

<div class="dd">

<code>tos</code>  <i>uint8</i>

</div>
<div class="dt">

Type of Service assigned to the class.

</div>

<hr />





## Path
Path represents a path to be probed

Appears in:


- <code><a href="#config">Config</a>.paths</code>





<hr />

<div class="dd">

<code>name</code>  <i>string</i>

</div>
<div class="dt">

Name for the path.

</div>

<hr />

<div class="dd">

<code>hops</code>  <i>[]string</i>

</div>
<div class="dt">

List of hops to probe.

</div>

<hr />

<div class="dd">

<code>measurement_length_ms</code>  <i>uint64</i>

</div>
<div class="dt">

Measurement interval expressed in milliseconds.

</div>

<hr />

<div class="dd">

<code>payload_size_bytes</code>  <i>uint64</i>

</div>
<div class="dt">

Payload size expressed in Bytes.

</div>

<hr />

<div class="dd">

<code>pps</code>  <i>uint64</i>

</div>
<div class="dt">

Amount of probing packets that will be sent per second.

</div>

<hr />

<div class="dd">

<code>timeout</code>  <i>uint64</i>

</div>
<div class="dt">

Timeout expressed in milliseconds.

</div>

<hr />





## Router
Router represents a router used a an explicit hop in a path

Appears in:


- <code><a href="#config">Config</a>.routers</code>





<hr />

<div class="dd">

<code>name</code>  <i>string</i>

</div>
<div class="dt">

Name of the router.

</div>

<hr />

<div class="dd">

<code>dst_range</code>  <i>string</i>

</div>
<div class="dt">

Destination range of IP addresses.

</div>

<hr />

<div class="dd">

<code>src_range</code>  <i>string</i>

</div>
<div class="dt">

Range of source ip addresses.

</div>

<hr />




