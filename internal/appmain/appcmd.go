// Package appmain contains the common application initialization code for Open Match servers.
package appmain

import (
	"github.com/sirupsen/logrus"
	"siody.home/om-like/internal/config"
	"siody.home/om-like/internal/logging"
)

// RunApplicationCmd starts and runs the given application cmd.  For use in
// main functions to run the full application.
func RunApplicationCmd(serviceName string, bindService Bind) {

	readConfig := func() (config.View, error) {
		return config.Read()
	}

	a, err := RunCmd(serviceName, bindService, readConfig)
	if err != nil {
		logger.Fatal(err)
	}

	err = a.Stop()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Application stopped successfully.")
}

// RunCmd is used internally, and public only for apptest.
func RunCmd(serviceName string, bindService Bind, getCfg func() (config.View, error)) (*App, error) {
	a := &App{}

	cfg, err := getCfg()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatalf("cannot read configuration.")
	}
	logging.ConfigureLogging(cfg)

	p := &Params{
		config:      cfg,
		serviceName: serviceName,
	}
	b := &Bindings{
		a: a,
	}

	err = bindService(p, b)
	if err != nil {
		surpressedErr := a.Stop() // Don't care about additional errors stopping.
		_ = surpressedErr
		return nil, err
	}
	if b.firstErr != nil {
		surpressedErr := a.Stop() // Don't care about additional errors stopping.
		_ = surpressedErr
		return nil, b.firstErr
	}

	return a, nil
}
