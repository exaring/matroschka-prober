package prober

import (
	"strings"
	"time"

	"github.com/exaring/matroschka-prober/pkg/measurement"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const (
	metricPrefix = "matroschka_"
)

// Describe is required by prometheus interface
func (p *Prober) Describe(ch chan<- *prometheus.Desc) {
}

// Collect collects data from the collector and send it to prometheus
func (p *Prober) Collect(ch chan<- prometheus.Metric) {
	ts := p.lastFinishedMeasurement()
	m := p.measurements.Get(ts)
	if m == nil {
		log.Infof("Requested timestamp %d not found", ts)
		return
	}

	p.collectSent(ch, m)
	p.collectReceived(ch, m)
	p.collectRTTMin(ch, m)
	p.collectRTTMax(ch, m)
	p.collectRTTAvg(ch, m)
}

func (p *Prober) labels() []string {
	keys := make([]string, len(p.staticLabels)+2)
	for i, l := range p.staticLabels {
		keys[i] = l.Key
	}

	keys[len(keys)-2] = "tos"
	keys[len(keys)-1] = "path"
	return keys
}

func (p *Prober) labelValues() []string {
	values := make([]string, len(p.staticLabels)+2)
	for i, l := range p.staticLabels {
		values[i] = l.Value
	}

	values[len(values)-2] = p.tos.LabelValue
	values[len(values)-1] = strings.Join(p.path.Hops, "-")
	return values
}

func (p *Prober) collectSent(ch chan<- prometheus.Metric, m *measurement.Measurement) {
	desc := prometheus.NewDesc(metricPrefix+"packets_sent", "Sent packets", p.labels(), nil)
	ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(m.Sent), p.labelValues()...)
}

func (p *Prober) collectReceived(ch chan<- prometheus.Metric, m *measurement.Measurement) {
	desc := prometheus.NewDesc(metricPrefix+"packets_received", "Received packets", p.labels(), nil)
	ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(m.Received), p.labelValues()...)
}

func (p *Prober) collectRTTMin(ch chan<- prometheus.Metric, m *measurement.Measurement) {
	desc := prometheus.NewDesc(metricPrefix+"rtt_min", "RTT Min", p.labels(), nil)
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(m.RTTMin), p.labelValues()...)
}

func (p *Prober) collectRTTMax(ch chan<- prometheus.Metric, m *measurement.Measurement) {
	desc := prometheus.NewDesc(metricPrefix+"rtt_max", "RTT Max", p.labels(), nil)
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(m.RTTMax), p.labelValues()...)
}

func (p *Prober) collectRTTAvg(ch chan<- prometheus.Metric, m *measurement.Measurement) {
	desc := prometheus.NewDesc(metricPrefix+"rtt_avg", "RTT Average", p.labels(), nil)
	v := float64(0)
	if m.Received != 0 {
		v = float64(m.RTTSum / m.Received)
	}
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, p.labelValues()...)
}

func (p *Prober) lastFinishedMeasurement() int64 {
	measurementLengthNS := int64(*p.path.MeasurementLengthMS) * int64(time.Millisecond)
	timeoutNS := int64(*p.path.TimeoutMS) * int64(time.Millisecond)
	nowNS := p.clock.Now().UnixNano()
	ts := nowNS - timeoutNS - measurementLengthNS
	return ts - ts%measurementLengthNS
}
