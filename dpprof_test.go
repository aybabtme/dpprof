package dpprof

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/http/pprof"
	"testing"
	"time"
)

func TestCanFetchProfile(t *testing.T) {
	count := 1
	var hosts []string
	for i := 0; i < count; i++ {
		srv := debugServer()
		defer srv.Close()
		hosts = append(hosts, srv.URL)
	}

	buf := bytes.NewBuffer(nil)

	err := CPU(buf, 2*time.Second, hosts...)
	if err != nil {
		t.Fatalf("failed to fetch CPU: %v", err)
	}
	t.Error("hello: \n" + buf.String())
}

func TestCanFetchCmdline(t *testing.T) {
	count := 10
	var hosts []string
	for i := 0; i < count; i++ {
		srv := debugServer()
		defer srv.Close()
		hosts = append(hosts, srv.URL)
	}

	dumps, err := Cmdline(hosts...)
	if err != nil {
		t.Fatalf("failed to fetch Cmdline: %v", err)
	}

	for _, data := range dumps {
		t.Error("hello: \n" + hex.Dump(data))
	}
}

func debugServer() *httptest.Server {
	return httptest.NewServer(debugHandler())
}

func debugHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	return mux
}
