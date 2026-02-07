package kite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sllt/kite/pkg/kite/infra"
	kiteHTTP "github.com/sllt/kite/pkg/kite/http"
	"github.com/sllt/kite/pkg/kite/http/middleware"
	"github.com/sllt/kite/pkg/kite/websocket"
)

type httpServer struct {
	router      *kiteHTTP.Router
	registry    *RouteRegistry
	port        int
	ws          *websocket.Manager
	srv         *http.Server
	certFile    string
	keyFile     string
	staticFiles map[string]string
}

var (
	errInvalidCertificateFile = errors.New("invalid certificate file")
	errInvalidKeyFile         = errors.New("invalid key file")
)

func newHTTPServer(c *infra.Container, port int, middlewareConfigs middleware.Config) *httpServer {
	r := kiteHTTP.NewRouter()
	wsManager := websocket.New()

	r.Use(
		middleware.Tracer,
		middleware.Logging(middlewareConfigs.LogProbes, c.Logger),
		middleware.CORS(middlewareConfigs.CorsHeaders, r.RegisteredRoutes),
		middleware.Metrics(c.Metrics()),
		middleware.WSHandlerUpgrade(c, wsManager),
	)

	return &httpServer{
		router:      r,
		registry:    newRouteRegistry(),
		port:        port,
		ws:          wsManager,
		staticFiles: make(map[string]string),
	}
}

func (s *httpServer) run(c *infra.Container) {
	if s.srv != nil {
		c.Logf("Server already running on port: %d", s.port)
		return
	}

	c.Logf("Starting server on port: %d", s.port)

	s.srv = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           s.router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// If both certFile and keyFile are provided, validate and run HTTPS server
	if s.certFile != "" && s.keyFile != "" {
		if err := validateCertificateAndKeyFiles(s.certFile, s.keyFile); err != nil {
			c.Error(err)
			return
		}

		// Start HTTPS server with TLS
		if err := s.srv.ListenAndServeTLS(s.certFile, s.keyFile); err != nil {
			c.Errorf("error while listening to https server, err: %v", err)
		}

		return
	}

	// If no certFile/keyFile is provided, run the HTTP server
	if err := s.srv.ListenAndServe(); err != nil {
		c.Errorf("error while listening to http server, err: %v", err)
	}
}

func (s *httpServer) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}

	return ShutdownWithContext(ctx, func(ctx context.Context) error {
		return s.srv.Shutdown(ctx)
	}, func() error {
		if err := s.srv.Close(); err != nil {
			return err
		}

		return nil
	})
}

func validateCertificateAndKeyFiles(certificateFile, keyFile string) error {
	if _, err := os.Stat(certificateFile); os.IsNotExist(err) {
		return fmt.Errorf("%w : %v", errInvalidCertificateFile, certificateFile)
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return fmt.Errorf("%w : %v", errInvalidKeyFile, keyFile)
	}

	return nil
}
