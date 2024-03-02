package main

import (
	"log"
	"net/http"
	"os"

	"roob.re/mastodon-api-proxy/proxy"
)

func main() {
	proxy, err := proxy.New(os.Getenv("BACKEND_URL"), os.Getenv("DOMAIN"))
	if err != nil {
		log.Fatalf("building proxy: %v", err)
	}

	const addr = ":8080"
	log.Printf("Starting server on %s", addr)
	err = http.ListenAndServe(addr, proxy)
	if err != nil {
		log.Fatalf("running server: %v", err)
	}
}
