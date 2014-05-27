package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/evanj/gohttpsserver"
)

func main() {
	port := flag.String("port", "8000", "Port to listen on")
	flag.Parse()

	log.Printf("Serving at https://localhost:%s/", *port)
	err := gohttpsserver.ServeWithNewSelfSigned(":"+*port, http.FileServer(http.Dir(".")))
	if err != nil {
		log.Fatal("failed to serve:", err)
	}
}
