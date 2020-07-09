package telemetry

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"siody.home/om-like/internal/config"
)

var (
	logger = logrus.WithFields(logrus.Fields{
		"app":       "openmatch",
		"component": "telemetry",
	})
)

// Setup configures the telemetry for the server.
func Setup(p Params, b Bindings) error {
	bindings := []func(p Params, b Bindings) error{
		configureOpenCensus,
		//bindJaeger,
		//bindPrometheus,
		//bindStackDriverMetrics,
		//bindOpenCensusAgent,
		//bindZpages,
		//bindHelp,
		//bindConfigz,
	}

	for _, f := range bindings {
		err := f(p, b)
		if err != nil {
			return err
		}
	}

	return nil
}

func configureOpenCensus(p Params, b Bindings) error {
	// There's no way to undo these options, but the next startup will override
	// them.

	samplingFraction := p.Config().GetFloat64("telemetry.traceSamplingFraction")
	logger.WithFields(logrus.Fields{
		"samplingFraction": samplingFraction,
	}).Info("Tracing sampler fraction set")
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(samplingFraction)})

	periodString := p.Config().GetString("telemetry.reportingPeriod")
	reportingPeriod, err := time.ParseDuration(periodString)
	if err != nil {
		return errors.Wrap(err, "Unable to parse telemetry.reportingPeriod")
	}
	logger.WithFields(logrus.Fields{
		"reportingPeriod": reportingPeriod,
	}).Info("Telemetry reporting period set")
	// Change the frequency of updates to the metrics endpoint
	view.SetReportingPeriod(reportingPeriod)
	return nil
}

// Params allows appmain to bind telemetry without a circular dependency.
type Params interface {
	Config() config.View
	ServiceName() string
}

// Bindings allows appmain to bind telemetry without a circular dependency.
type Bindings interface {
	TelemetryHandle(pattern string, handler http.Handler)
	TelemetryHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	AddCloser(c func())
	AddCloserErr(c func() error)
}
