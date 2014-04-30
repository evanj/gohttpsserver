package main

import (
	"flag"
	"fmt"
	"log"
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

	proxy := gohttpsserver.NewSingleHostReverseProxy(remote)

	log.Printf("Serving at https://localhost:%d/", *port)
	err = gohttpsserver.ServeWithGeneratedCert(":"+strconv.Itoa(*port), proxy)
	if err != nil {
		log.Fatal("failed to serve: ", err)
	}
}
