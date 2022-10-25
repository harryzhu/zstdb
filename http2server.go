package sqlconf

import (
	"os"
	//"io/ioutil"
	//"mime"
	"net/http"
	//"os"

	//"path/filepath"
	"strconv"
	"strings"

	//"time"
	//"fmt"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type H2Server struct {
	StaticRootDir string
	IP            string
	Port          int
	TLScert       string
	TLSkey        string
	AllowIPList   []string
	BlockIPList   []string
	DefaultAllow  bool
}

var h2server *H2Server = &H2Server{
	StaticRootDir: "./",
	IP:            "0.0.0.0",
	Port:          8080,
	TLScert:       "./cert.pem",
	TLSkey:        "./priv.key",
	AllowIPList:   []string{},
	BlockIPList:   []string{},
	DefaultAllow:  true,
}

func (h2s *H2Server) WithStaticRootDir(s string) *H2Server {
	h2s.StaticRootDir = s
	return h2s
}

func (h2s *H2Server) WithIP(s string) *H2Server {
	h2s.IP = s
	return h2s
}

func (h2s *H2Server) WithPort(i int) *H2Server {
	h2s.Port = i
	return h2s
}

func (h2s *H2Server) WithTLS(c, k string) *H2Server {
	h2s.TLScert = c
	h2s.TLSkey = k
	return h2s
}

func (h2s *H2Server) WithAllowIP(s string) *H2Server {
	h2s.AllowIPList = append(h2s.AllowIPList, s)
	return h2s
}

func (h2s *H2Server) runH2Server() {

	if h2s.StaticRootDir == "" {
		h2s.StaticRootDir = "./"
	}

	if h2s.Port <= 0 {
		h2s.Port = 8080
	}

	if h2s.TLScert == "" {
		zapLogger.Error("you have to set a trusted cert")
	}

	if h2s.TLSkey == "" {
		zapLogger.Error("you have to set a trusted key")
	}

	if _, err := os.Stat(h2s.TLScert); err != nil {
		zapLogger.Error("h2s.TLScert does not exist", zap.Error(err))
	}

	if _, err := os.Stat(h2s.TLSkey); err != nil {
		zapLogger.Error("h2s.TLSkey does not exist", zap.Error(err))
	}

	addr := strings.Join([]string{h2s.IP, strconv.Itoa(h2s.Port)}, ":")
	server := http.Server{
		Addr:    addr,
		Handler: http.FileServer(http.Dir(h2s.StaticRootDir)),
	}

	visitURL := "https://your-domain-same-as-your-cert-key:" + strconv.Itoa(h2s.Port) + "/"

	zapLogger.Info("http2 server",
		zap.String("StaticRootDir", h2s.StaticRootDir),
		zap.String("Address", visitURL),
		zap.String("TLScert", h2s.TLScert),
		zap.String("TLSkey", h2s.TLSkey),
	)

	http2.ConfigureServer(&server, &http2.Server{})

	err := server.ListenAndServeTLS(h2s.TLScert, h2s.TLSkey)
	if err != nil {
		zapLogger.Error("runControlServer", zap.Error(err))
	}
}

func (h2s *H2Server) StartServer() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		h2s.runH2Server()
	}()

	go func() {
		h2s.runControlServer()
	}()

	wg.Wait()
}
