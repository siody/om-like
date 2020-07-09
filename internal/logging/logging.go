// Package logging configures the Logrus logging library.
package logging

import (
	"strings"

	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
	"siody.home/om-like/internal/config"
)

// ConfigureLogging sets up open match logrus instance using the logging section of the matchmaker_config.json
//  - log line format (text[default] or json)
//  - min log level to include (debug, info [default], warn, error, fatal, panic)
func ConfigureLogging(cfg config.View) {
	logrus.SetFormatter(newFormatter(cfg.GetString("logging.format")))
	level := toLevel(cfg.GetString("logging.level"))
	logrus.SetLevel(level)
	if isDebugLevel(level) {
		logrus.Warn("Trace logging level configured. Not recommended for production!")
	}
}

func newFormatter(formatter string) logrus.Formatter {
	switch strings.ToLower(formatter) {
	case "stackdriver":
		return stackdriver.NewFormatter()
	case "json":
		return &logrus.JSONFormatter{}
	}
	return &logrus.TextFormatter{}
}

func toLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "warn":
		fallthrough
	case "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	}
	return logrus.InfoLevel
}

// IsDebugEnabled returns true if the logging level is debug or more granular.
func IsDebugEnabled(cfg config.View) bool {
	return IsDebugLevel(cfg.GetString("logging.level"))
}

// IsDebugLevel returns true if the logging level is debug or more granular.
func IsDebugLevel(level string) bool {
	return isDebugLevel(toLevel(level))
}

func isDebugLevel(level logrus.Level) bool {
	switch level {
	case logrus.TraceLevel:
		fallthrough
	case logrus.DebugLevel:
		return true
	}
	return false
}
