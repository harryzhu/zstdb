package sqlconf

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	//"fmt"
	//"go.uber.org/zap"
	//"golang.org/x/net/http2"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if r.URL.Path == "/" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
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
		zapLogger.Info("app will exit")
		os.Exit(0)
	}()

}

func (h2s *H2Server) runControlServer() {
	addr := strings.Join([]string{h2s.IP, strconv.Itoa(h2s.Port + 1)}, ":")

	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/remote-shutdown", RemoteShutdownHandler)

	http.ListenAndServe(addr, mux)
}
