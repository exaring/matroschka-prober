package frontend

import (
	"net/http"

	"github.com/exaring/matroschka-prober/pkg/config"
	"github.com/exaring/matroschka-prober/pkg/prober"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promlog "github.com/prometheus/common/log"
	log "github.com/sirupsen/logrus"
)

// Frontend represents an HTTP prometheus interface
type Frontend struct {
	cfg        *config.Config
	collectors []*prober.Prober
}

// New creates a new HTTP frontend
func New(cfg *config.Config, collectors []*prober.Prober) *Frontend {
	return &Frontend{
		cfg:        cfg,
		collectors: collectors,
	}
}

// Start starts the frontend
func (fe *Frontend) Start() {
	log.Infof("Starting Matroschka Prober (Version: %s)\n", fe.cfg.Version)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Matroschka Prober (Version ` + fe.cfg.Version + `)</title></head>
			<body>
			<h1>Matroschka Prober</h1>
			<p><a href="` + *fe.cfg.MetricsPath + `">Metrics</a></p>
			<h2>More information:</h2>
			<p><a href="https://github.com/exaring/matroschka-prober">github.com/exaring/matroschka-prober</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc(*fe.cfg.MetricsPath, fe.handleMetricsRequest)

	log.Infof("Listening for %s on %s\n", *fe.cfg.MetricsPath, *fe.cfg.ListenAddress)
	log.Fatal(http.ListenAndServe(*fe.cfg.ListenAddress, nil))
}

func (fe *Frontend) handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	reg := prometheus.NewRegistry()
	for i := range fe.collectors {
		reg.MustRegister(fe.collectors[i])
	}

	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:      promlog.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, r)
}
