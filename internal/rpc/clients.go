package rpc

import (
	"github.com/sirupsen/logrus"
)

const (
	// ConfigNameEnableRPCLogging is the config name for enabling RPC logging.
	ConfigNameEnableRPCLogging = "logging.rpc"
	// configNameClientTrustedCertificatePath is the same as the root CA cert that the server trusts.
	configNameClientTrustedCertificatePath = configNameServerRootCertificatePath
)

var (
	clientLogger = logrus.WithFields(logrus.Fields{
		"app":       "openmatch",
		"component": "client",
	})
)
