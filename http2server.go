package sqlconf

import (
	"net/url"
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
	StaticRootDir   string
	IP              string
	Port            int
	TLScert         string
	TLSkey          string
	DefaultAllow    bool
	EnableControl   bool
	EnableProxy     bool
	ReverseProxyURL string
}

var AllowBlockIPMap map[string]int = make(map[string]int, 32)

var h2server *H2Server = &H2Server{
	StaticRootDir:   "./",
	IP:              "0.0.0.0",
	Port:            8080,
	TLScert:         "./cert.pem",
	TLSkey:          "./priv.key",
	DefaultAllow:    true,
	EnableControl:   false,
	EnableProxy:     false,
	ReverseProxyURL: "",
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
	AllowBlockIPMap[s] = 1
	return h2s
}

func (h2s *H2Server) WithBlockIP(s string) *H2Server {
	AllowBlockIPMap[s] = -1
	return h2s
}

func (h2s *H2Server) WithDefaultAllow(b bool) *H2Server {
	h2s.DefaultAllow = b
	return h2s
}

func IsAllow(ipaddr string) bool {
	var ipint int = 0
	if _, ok := AllowBlockIPMap[ipaddr]; ok {
		ipint = AllowBlockIPMap[ipaddr]
	}

	if h2server.DefaultAllow == true {
		if ipint == -1 {
			return false
		}
		return true
	} else {
		if ipint == 1 {
			return true
		}
		return false
	}
}

type staticHandler struct {
	rootDir string
}

func (sh staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cip := GetClientIP(r)
	isa := IsAllow(cip)
	if isa != true {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(cip + ", " + r.Proto + " ,you cannot visit this site."))
		return
	}

	http.FileServer(http.Dir(sh.rootDir)).ServeHTTP(w, r)
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
	// server := http.Server{
	// 	Addr:    addr,
	// 	Handler: http.FileServer(http.Dir(h2s.StaticRootDir)),
	// }

	server := http.Server{
		Addr:    addr,
		Handler: &staticHandler{rootDir: h2s.StaticRootDir},
	}

	visitURL := "https://your-domain-same-as-your-cert-key:" + strconv.Itoa(h2s.Port) + "/"
	var allowIPList []string
	var blockIPList []string
	for k, v := range AllowBlockIPMap {
		if v == 1 {
			allowIPList = append(allowIPList, k)
		}

		if v == -1 {
			blockIPList = append(blockIPList, k)
		}
	}
	allowiplist := strings.Join(allowIPList, ";")
	blockiplist := strings.Join(blockIPList, ";")

	zapLogger.Info("http2 server",
		zap.String("StaticRootDir", h2s.StaticRootDir),
		zap.String("Address", visitURL),
		zap.String("TLScert", h2s.TLScert),
		zap.String("TLSkey", h2s.TLSkey),
		zap.String("AllowIPList", allowiplist),
		zap.String("BlockIPList", blockiplist),
		zap.Bool("DefaultAllow", h2s.DefaultAllow),
	)

	http2.ConfigureServer(&server, &http2.Server{})

	zapLogger.Info("runH2Server", zap.String("address", addr))
	err := server.ListenAndServeTLS(h2s.TLScert, h2s.TLSkey)
	if err != nil {
		zapLogger.Error("runH2Server", zap.Error(err))
	}
}

func (h2s *H2Server) runReverseProxy() error {
	if h2s.ReverseProxyURL == "" {
		zapLogger.Error("ReverseProxyURL cannot be empty")
		return nil
	}
	remote, err := url.Parse(h2s.ReverseProxyURL)
	if err != nil {
		zapLogger.Error("runReverseProxy", zap.String("reverseUrl", h2s.ReverseProxyURL), zap.Error(err))
		return err
	}

	proxy := GoReverseProxy(&RProxy{
		remote: remote,
	})
	addr := strings.Join([]string{h2s.IP, strconv.Itoa(h2s.Port + 2)}, ":")

	zapLogger.Info("runReverseProxy", zap.String("address", addr))
	err = http.ListenAndServeTLS(addr, h2s.TLScert, h2s.TLSkey, proxy)

	if err != nil {
		zapLogger.Error("runReverseProxy", zap.Error(err))
		return err
	}

	return nil
}

func (h2s *H2Server) StartServer() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		h2s.runH2Server()
	}()

	if h2s.EnableControl == true {
		wg.Add(1)
		go func() {
			h2s.runControlServer()
		}()
	}

	if h2s.EnableProxy == true {
		wg.Add(1)
		go func() {
			h2s.runReverseProxy()
		}()
	}

	wg.Wait()
}
