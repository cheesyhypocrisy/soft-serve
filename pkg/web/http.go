package web

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/soft-serve/pkg/config"
)

// HTTPServer is an http server.
type HTTPServer struct {
	ctx context.Context
	cfg *config.Config
	cr  *CertReloader

	Server *http.Server
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(ctx context.Context) (*HTTPServer, error) {
	cfg := config.FromContext(ctx)
	logger := log.FromContext(ctx)
	s := &HTTPServer{
		ctx: ctx,
		cfg: cfg,
		Server: &http.Server{
			Addr:              cfg.HTTP.ListenAddr,
			Handler:           NewRouter(ctx),
			ReadHeaderTimeout: time.Second * 10,
			IdleTimeout:       time.Second * 10,
			MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
			ErrorLog:          logger.StandardLog(log.StandardLogOptions{ForceLevel: log.ErrorLevel}),
		},
	}

	if cfg.HTTP.TLSKeyPath != "" && cfg.HTTP.TLSCertPath != "" {
		cr, err := NewCertReloader(cfg.HTTP.TLSCertPath, cfg.HTTP.TLSKeyPath, logger)
		if err != nil {
			return nil, err
		}
		s.cr = cr

		s.Server.TLSConfig = &tls.Config{
			GetCertificate: cr.GetCertificateFunc(),
		}
	}

	return s, nil
}

// Close closes the HTTP server.
func (s *HTTPServer) Close() error {
	return s.Server.Close()
}

// ListenAndServe starts the HTTP server.
func (s *HTTPServer) ListenAndServe() error {
	if s.Server.TLSConfig != nil {
		return s.Server.ListenAndServeTLS("", "")
	}
	return s.Server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
