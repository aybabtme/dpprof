package main

import (
	"compress/gzip"
	"crypto/rand"
	"flag"
	"github.com/aybabtme/iocontrol"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	port := flag.String("port", "8080", "port to listen on")
	flag.Parse()
	log.Printf("port %q", *port)

	go func() {
		m := iocontrol.ThrottledReader(rand.Reader, iocontrol.KiB*100, time.Millisecond*500)
		io.Copy(gzip.NewWriter(ioutil.Discard), m)
	}()

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
