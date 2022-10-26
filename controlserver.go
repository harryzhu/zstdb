package sqlconf

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	//"fmt"
	"go.uber.org/zap"
	//"golang.org/x/net/http2"
)

func GetClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	cip := GetClientIP(r)
	zapLogger.Info("client", zap.String("ip", cip))
	zapLogger.Info("allow/block", zap.Bool("is-allow", IsAllow(cip)))
	if IsAllow(cip) != true {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(cip + ",you cannot visit this site.<br/>"))
		w.Write([]byte(r.Proto))
		return
	}
	if r.URL.Path == "/" {
		w.WriteHeader(http.StatusOK)

		w.Write([]byte("welcome<br/>"))
		w.Write([]byte(cip + "<br/>"))
		w.Write([]byte(r.Proto))
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	}

}

func RemoteShutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusGone)
	w.Write([]byte("shutdown the server in 3 seconds..."))
	go func() {
		time.Sleep(3 * time.Second)
		zapLogger.Info("app will exit after 3 seconds")
		os.Exit(0)
	}()

}

func (h2s *H2Server) runControlServer() {
	addr := strings.Join([]string{h2s.IP, strconv.Itoa(h2s.Port + 1)}, ":")

	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/remote-shutdown", RemoteShutdownHandler)

	err := http.ListenAndServeTLS(addr, h2s.TLScert, h2s.TLSkey, mux)
	if err != nil {
		zapLogger.Error("runControlServer", zap.Error(err))
	}
}
