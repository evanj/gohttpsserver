package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/evanj/gohttpsserver"
)

func makeProxyHeaderDirector(originalDirector func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		log.Printf("%s //%s%s %s", r.Method, r.Host, r.URL, r.RemoteAddr)
		// X-Forwarded-For is set be ReverseProxy; X-Forwarded-Proto is not
		protocol := "http"
		if r.TLS != nil {
			protocol = "https"
		}
		r.Header.Set("X-Forwarded-Proto", protocol)
		originalDirector(r)
	}
}

func fatalCommand(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		fatalCommand("Usage: https_proxy (remote)")
		return
	}
	remote, err := url.Parse(os.Args[1])
	if err != nil {
		fatalCommand("Invalid URL:", err)
		return
	}
	// Early check for errors rather than at connect time
	// TODO: Error check by calling proxy.Transport.RoundTrip instead?
	if remote.Scheme != "http" && remote.Scheme != "https" {
		fatalCommand("Unsupported scheme:", remote.Scheme)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = makeProxyHeaderDirector(proxy.Director)

	log.Print("Serving at https://localhost:8001/")
	err = gohttpsserver.ServeWithGeneratedCert(":8001", proxy)
	if err != nil {
		log.Fatal("failed to serve:", err)
	}
}
