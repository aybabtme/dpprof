package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aybabtme/dpprof"
)

func main() {
	hosts := ListFlag{}
	flag.Var(&hosts, "h", "enumerated list of hosts, invoked once for each host")
	addr := flag.String("addr", ":8000", "interface + port to listen on")
	flag.Parse()

	if len(hosts) < 1 {
		log.Fatal("need at least one host")
	}

	log.Printf("listening on %q, proxying to %q", *addr, hosts)

	prof := &Profiler{hosts: hosts}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.NewBuffer(nil)
		_, _ = io.Copy(buf, r.Body)
		log.Printf("not found: %#v, %q", r, r.URL.String())
		log.Printf("%s", buf.String())
	})
	mux.HandleFunc("/debug/pprof/", prof.NamedProfile)
	mux.HandleFunc("/debug/pprof/profile", prof.CPU)
	mux.HandleFunc("/debug/pprof/symbol", prof.Symbol)
	log.Fatal(http.ListenAndServe(*addr, mux))

}

type Profiler struct {
	hosts []string
}

func (p *Profiler) CPU(w http.ResponseWriter, r *http.Request) {

	log.Printf("proxy for %q to %v", r.URL.String(), p.hosts)

	buf := bytes.NewBuffer(nil)

	err := dpprof.CPU(buf, 30*time.Second, p.hosts...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = buf.WriteTo(w)
}

func (p *Profiler) Symbol(w http.ResponseWriter, r *http.Request) {

	log.Printf("proxy for %q to %v", r.URL.String(), p.hosts)

	resp, err := dpprof.Symbol(r.Body, p.hosts...)
	if err != nil {
		log.Printf("err=%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.Copy(w, resp)
}

func (p *Profiler) NamedProfile(w http.ResponseWriter, r *http.Request) {

	log.Printf("proxy for %q to %v", r.URL.String(), p.hosts)

	code := 200
	buf := bytes.NewBuffer(nil)
	err := dpprof.NamedProfile(buf, r, p.hosts...)
	if e, ok := err.(*dpprof.Error); ok {
		code = e.Code
		buf.WriteString(e.Msg)
	} else if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), code)
		return
	}
	w.WriteHeader(code)
	_, _ = buf.WriteTo(w)
}
