package rpc

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/pkg/errors"

	"siody.home/om-like/internal/telemetry"
)

const (
	// https://http2.github.io/http2-spec/#rfc.section.3.1
	http2WithTLSVersionID = "h2"
)

type tlsServer struct {
	httpListener net.Listener
	httpMux      *http.ServeMux
	httpServer   *http.Server
}

func (s *tlsServer) start(params *ServerParams) error {
	s.httpMux = params.ServeMux

	rootCaCert, err := trustedCertificateFromFileData(params.rootCaPublicCertificateFileData)
	if err != nil {
		return errors.WithStack(err)
	}

	tlsCertificate, err := certificateFromFileData(params.publicCertificateFileData, params.privateKeyFileData)
	if err != nil {
		return errors.WithStack(err)
	}

	// Start HTTP server
	for _, handlerFunc := range params.handlerForHTTP {
		handlerFunc(s.httpMux)
	}

	// Bind HTTPS handlers
	s.httpMux.Handle(telemetry.HealthCheckEndpoint, telemetry.NewHealthCheck(params.handlersForHealthCheck))
	s.httpServer = &http.Server{
		Addr:    s.httpListener.Addr().String(),
		Handler: instrumentHTTPHandler(s.httpMux, params),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*tlsCertificate},
			ClientCAs:    rootCaCert,
			// Don`t support mutual authentication yet
			//ClientAuth:   tls.RequireAndVerifyClientCert,
			NextProtos: []string{http2WithTLSVersionID},
		},
	}
	go func() {
		tlsListener := tls.NewListener(s.httpListener, s.httpServer.TLSConfig)
		serverLogger.Infof("Serving HTTPS: %s", s.httpListener.Addr().String())
		hErr := s.httpServer.Serve(tlsListener)
		if hErr != nil && hErr != http.ErrServerClosed {
			serverLogger.Debugf("error serving HTTP: %s", hErr)
		}
	}()

	return nil
}

func (s *tlsServer) stop() error {
	// the servers also close their respective listeners.
	err := s.httpServer.Shutdown(context.Background())
	return err
}

func newTLSServer(httpL net.Listener) *tlsServer {
	return &tlsServer{
		httpListener: httpL,
	}
}
