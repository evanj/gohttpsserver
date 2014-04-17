package main

import (
	"log"
	"net/http"

	"github.com/evanj/gohttpsserver"
)

func main() {
	log.Print("Serving at https://localhost:8000/")
	err := gohttpsserver.ServeWithGeneratedCert(":8000", http.FileServer(http.Dir(".")))
	if err != nil {
		log.Fatal("failed to serve:", err)
	}
}
