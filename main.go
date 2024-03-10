package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"roob.re/mastodon-api-proxy/proxy"
)

func main() {
	log := slog.Default()
	opts := proxy.Options{
		Logger: log,
	}

	minAgeStr := os.Getenv("MIN_AGE")
	if minAgeStr != "" {
		minAge, err := time.ParseDuration(minAgeStr)
		if err != nil {
			log.With("MIN_AGE", minAgeStr).Error("parsing MIN_AGE")
			return
		}

		log.With("minAge", minAge).Info("Restricting account age")
		opts.MinAge = minAge
	}

	proxy, err := proxy.New(
		os.Getenv("BACKEND_URL"),
		os.Getenv("DOMAIN"),
		opts,
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
