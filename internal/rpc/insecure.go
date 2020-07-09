package rpc

import (
	"context"
	"net"
	"net/http"
	"siody.home/om-like/internal/telemetry"
)

type insecureServer struct {


	httpListener net.Listener
	httpMux      *http.ServeMux
	httpServer   *http.Server
}

func (s *insecureServer) start(params *ServerParams) error {
	s.httpMux = params.ServeMux

	// Configure the HTTP proxy server.
	// Bind gRPC handlers

	for _, handlerFunc := range params.handlerForHttp {
		handlerFunc(s.httpMux)
	}

	s.httpMux.Handle(telemetry.HealthCheckEndpoint, telemetry.NewHealthCheck(params.handlersForHealthCheck))
	s.httpServer = &http.Server{
		Addr:    s.httpListener.Addr().String(),
		Handler: instrumentHTTPHandler(s.httpMux, params),
	}
	go func() {
		serverLogger.Infof("Serving HTTP: %s", s.httpListener.Addr().String())
		hErr := s.httpServer.Serve(s.httpListener)
		if hErr != nil && hErr != http.ErrServerClosed {
			serverLogger.Debugf("error serving HTTP: %s", hErr)
		}
	}()

	return nil
}

func (s *insecureServer) stop() error {
	// the servers also close their respective listeners.
	err := s.httpServer.Shutdown(context.Background())
	return err
}

func newInsecureServer(httpL net.Listener) *insecureServer {
	return &insecureServer{
		httpListener: httpL,
	}
}
