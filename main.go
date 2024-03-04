package main

import (
	"log/slog"
	"net/http"
	"os"

	"roob.re/mastodon-api-proxy/proxy"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	proxy, err := proxy.New(
		log,
		os.Getenv("BACKEND_URL"),
		os.Getenv("DOMAIN"),
	)
	if err != nil {
		log.Error("building proxy", "error", err)
		return
	}

	const addr = ":8080"
	log.Info("Starting server", "address", addr)

	err = http.ListenAndServe(addr, proxy)
	if err != nil {
		log.Error("running server", "error", err)
		return
	}
}
