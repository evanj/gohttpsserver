package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/evanj/gohttpsserver"
)

func fatalCommand(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

func main() {
	port := flag.Int("port", 8001, "port to listen on")
	disableCertificateValidation := flag.Bool("disableCertificateValidation", false, "DANGEROUS: ignore SSL errors on outgoing connections")
	host := flag.String("host", "", "value to override the Host header")
	mapping := flag.String("mapping", "", "additional space-separated (prefix, target) mappings")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fatalCommand("Usage: https_proxy (remote)")
		return
	}
	remote, err := url.Parse(flag.Arg(0))
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

	var transport http.RoundTripper
	if *disableCertificateValidation {
		defaultCopy := *http.DefaultTransport.(*http.Transport)
		defaultCopy.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		transport = &defaultCopy
	}

	proxy := gohttpsserver.NewSingleHostReverseProxy(remote)
	proxy.Transport = transport
	proxy.OverrideHost = *host

	mappings := gohttpsserver.ParseMappings(*mapping)
	for _, mapping := range mappings {
		proxy.MapPrefix(mapping.Prefix, mapping.Target)
	}

	log.Printf("Serving at https://localhost:%d/", *port)
	err = gohttpsserver.ServeWithNewSelfSigned(":"+strconv.Itoa(*port), proxy)
	if err != nil {
		log.Fatal("failed to serve: ", err)
	}
}
