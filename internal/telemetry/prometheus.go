package telemetry

// Taken from https://opencensus.io/quickstart/go/metrics/#1
import (
	ocPrometheus "contrib.go.opencensus.io/exporter/prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/stats/view"
)

const (
	// ConfigNameEnableMetrics indicates that telemetry is enabled.
	ConfigNameEnableMetrics = "telemetry.prometheus.enable"
)

func bindPrometheus(p Params, b Bindings) error {
	cfg := p.Config()

	if !cfg.GetBool("telemetry.prometheus.enable") {
		logger.Info("Prometheus Metrics: Disabled")
		return nil
	}

	endpoint := cfg.GetString("telemetry.prometheus.endpoint")

	logger.WithFields(logrus.Fields{
		"endpoint": endpoint,
	}).Info("Prometheus Metrics: ENABLED")

	registry := prometheus.NewRegistry()
	// Register standard prometheus instrumentation.
	err := registry.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	if err != nil {
		return errors.Wrap(err, "Failed to register prometheus collector")
	}
	err = registry.Register(prometheus.NewGoCollector())
	if err != nil {
		return errors.Wrap(err, "Failed to register prometheus collector")
	}

	promExporter, err := ocPrometheus.NewExporter(
		ocPrometheus.Options{
			Namespace: "",
			Registry:  registry,
		})
	if err != nil {
		return errors.Wrap(err, "Failed to initialize OpenCensus exporter to Prometheus")
	}

	// Register the Prometheus exporters as a stats exporter.
	view.RegisterExporter(promExporter)
	b.AddCloser(func() {
		view.UnregisterExporter(promExporter)
	})

	b.TelemetryHandle(endpoint, promExporter)
	return nil
}
