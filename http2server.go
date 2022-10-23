package sqlconf

import (
	"net/http"
	//"time"

	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type H2Server struct {
	StaticRootDir string
	Address       string
	TLScert       string
	TLSkey        string
}

var h2server *H2Server = &H2Server{}

func (h2s *H2Server) WithStaticRootDir(s string) *H2Server {
	h2s.StaticRootDir = s
	return h2s
}

func (h2s *H2Server) WithAddress(s string) *H2Server {
	h2s.Address = s
	return h2s
}

func (h2s *H2Server) WithTLS(c, k string) *H2Server {
	h2s.TLScert = c
	h2s.TLSkey = k
	return h2s
}

func (h2s *H2Server) StartServer() {

	if h2s.StaticRootDir == "" {
		h2s.StaticRootDir = "./"
	}

	if h2s.Address == "" {
		h2s.Address = ":8080"
	}

	if h2s.TLScert == "" {
		h2s.TLScert = "cert.pem"
	}

	if h2s.TLSkey == "" {
		h2s.TLSkey = "priv.key"
	}

	server := http.Server{
		Addr:    h2s.Address,
		Handler: http.FileServer(http.Dir(h2s.StaticRootDir)),
	}

	zapLogger.Info("http2 server",
		zap.String("StaticRootDir", h2s.StaticRootDir),
		zap.String("Address", h2s.Address),
		zap.String("TLScert", h2s.TLScert),
		zap.String("TLSkey", h2s.TLSkey),
	)

	http2.ConfigureServer(&server, &http2.Server{})
	server.ListenAndServeTLS(h2s.TLScert, h2s.TLSkey)
}
