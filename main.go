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

	if minAgeStr := os.Getenv("MIN_AGE"); minAgeStr != "" {
		minAge, err := time.ParseDuration(minAgeStr)
		if err != nil {
			log.With("MIN_AGE", minAgeStr).Error("parsing MIN_AGE")
			return
		}

		log.With("minAge", minAge).Info("Restricting account age")
		opts.MinAge = minAge
	}

	if maxStatusAgeStr := os.Getenv("MAX_STATUS_AGE"); maxStatusAgeStr != "" {
		maxStatusAge, err := time.ParseDuration(maxStatusAgeStr)
		if err != nil {
			log.With("MAX_STATUS_AGE", maxStatusAgeStr).Error("parsing MAX_STATUS_AGE")
			return
		}

		log.With("maxStatusAge", maxStatusAge).Info("Restricting max last status age")
		opts.MaxStatusAge = maxStatusAge
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
