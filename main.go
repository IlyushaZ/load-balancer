package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL   *url.URL
	Alive bool
	Proxy *httputil.ReverseProxy
	mux   sync.RWMutex
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return
}

type Pool struct {
	Backends []*Backend
	Current  uint64
}

func (p *Pool) Next() int {
	return int(atomic.AddUint64(&p.Current, 1) % uint64(len(p.Backends)))
}

func (p *Pool) PickBackend() *Backend {
	n := p.Next()
	l := n + len(p.Backends)

	for i := n; i < l; i++ {
		id := i % len(p.Backends)

		if p.Backends[id].IsAlive() {
			atomic.StoreUint64(&p.Current, uint64(id))
			return p.Backends[id]
		}
	}

	return nil
}

func Balance(w http.ResponseWriter, r *http.Request) {
	b := pool.PickBackend()
	if b != nil {
		b.Proxy.ServeHTTP(w, r)
		return
	}

	http.Error(w, "service is not available", http.StatusServiceUnavailable)
}

func HealthCheck() {
	for _, b := range pool.Backends {
		if !alive(b) {
			b.SetAlive(false)
			log.Println("backend is not available: ", b.URL.Host)
		}
	}
}

func alive(b *Backend) bool {
	conn, err := net.DialTimeout("tcp", b.URL.Host, 3*time.Second)
	if err != nil {
		return false
	}

	_ = conn.Close()
	return true
}

const (
	defaultFilename = "backends.txt"
	defaultSchema   = "http"
)

var pool Pool

func main() {
	var filename, schema string
	flag.StringVar(&filename, "f", defaultFilename, "Config files with ips to be used as backends")
	flag.StringVar(&schema, "s", defaultSchema, "Schema to be used (http/https)")
	flag.Parse()

	if schema != "http" && schema != "https" {
		log.Fatal("invalid schema entered")
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)

	for s.Scan() {
		u, err := url.Parse(schema + "://" + s.Text())
		if err != nil {
			continue
		}

		pool.Backends = append(pool.Backends, &Backend{
			URL:   u,
			Alive: true,
			Proxy: httputil.NewSingleHostReverseProxy(u),
		})
	}

	_ = f.Close()

	go func() {
		t := time.NewTicker(3 * time.Minute)

		for {
			select {
			case <-t.C:
				log.Println("started healthcheck")
				HealthCheck()
				log.Println("finished healthcheck")
			}
		}
	}()

	server := http.Server{
		Handler: http.HandlerFunc(Balance),
		Addr:    ":80",
	}

	log.Fatal(server.ListenAndServe())
}
