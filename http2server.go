package sqlconf

import (
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
}

var h2server *H2Server = &H2Server{}

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

// func (hh *h2sHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path == "/remote-shutdown/bbc" {
// 		fmt.Fprintf(w, "shutdown")
// 	} else {
// 		fullPath := filepath.Join(h2server.StaticRootDir, r.URL.Path)
// 		zapLogger.Info("file", zap.String("path", fullPath))
// 		_, err := os.Stat(fullPath)
// 		if err != nil {
// 			fmt.Fprintf(w, "404 page not found")
// 		}
// 		f, err := os.Open(fullPath)
// 		if err != nil {
// 			fmt.Fprintf(w, "404 page not found(open)")
// 		}
// 		defer f.Close()

// 		fd, err := ioutil.ReadFile(fullPath)
// 		if err != nil {
// 			fmt.Fprintf(w, "404 page not found(read)")
// 		}
// 		fmime := mime.TypeByExtension(filepath.Ext(fullPath))
// 		fmt.Println("mime:", fmime)
// 		w.Header().Set("Content-Type", fmime)

// 		w.Write(fd)
// 	}
// 	w.Header().Set("Content-Type", "text/html")
// 	fmt.Fprintf(w, "OK")
// }

func (h2s *H2Server) runH2Server() {

	if h2s.StaticRootDir == "" {
		h2s.StaticRootDir = "./"
	}

	if h2s.Port <= 0 {
		h2s.Port = 8080
	}

	if h2s.TLScert == "" {
		h2s.TLScert = "cert.pem"
	}

	if h2s.TLSkey == "" {
		h2s.TLSkey = "priv.key"
	}

	server := http.Server{
		Addr:    strings.Join([]string{h2s.IP, strconv.Itoa(h2s.Port)}, ":"),
		Handler: http.FileServer(http.Dir(h2s.StaticRootDir)),
	}

	zapLogger.Info("http2 server",
		zap.String("StaticRootDir", h2s.StaticRootDir),
		zap.Int("Port", h2s.Port),
		zap.String("TLScert", h2s.TLScert),
		zap.String("TLSkey", h2s.TLSkey),
	)

	http2.ConfigureServer(&server, &http2.Server{})

	server.ListenAndServeTLS(h2s.TLScert, h2s.TLSkey)

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
