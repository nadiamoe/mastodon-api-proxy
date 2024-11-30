package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	proxy  *httputil.ReverseProxy
	domain string
	opts   Options
}

type Options struct {
	Logger *slog.Logger
	MinAge time.Duration
}

func (o Options) defaults() Options {
	o.Logger = slog.Default()

	return o
}

func New(backendUrl, domain string, opts Options) (Handler, error) {
	opts = opts.defaults()

	u, err := url.Parse(backendUrl)
	if err != nil {
		return Handler{}, fmt.Errorf("parsing backend url: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	h := Handler{
		proxy:  proxy,
		domain: domain,
		opts:   opts,
	}

	const verifyCredentialsURL = "/api/v1/accounts/verify_credentials"

	proxy.ModifyResponse = func(r *http.Response) error {
		h.opts.Logger.With("path", r.Request.URL.Path, "upstream_status", r.Status).Info("handling response")

		switch r.Request.URL.Path {
		case verifyCredentialsURL:
			return h.addFakeEmail(r)
		default:
			return nil
		}
	}

	return h, nil
}

func (h Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.opts.Logger.With("path", r.URL.Path).Info("handling request")
	h.proxy.ServeHTTP(rw, r)
}

func (h Handler) addFakeEmail(r *http.Response) error {
	log := h.opts.Logger.With("path", r.Request.URL.Path, "upstream_status", r.Status)

	log.Info("processing verify_credentials response")

	if r.StatusCode != http.StatusOK {
		log.Warn("forwarding unmodified non-200 response")
		return nil
	}

	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("reading upstream body: %w", err)
	}

	var payload map[string]any
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error("cannot parse upstream body")
		return nil
	}

	const acctKey = "acct" // https://docs.joinmastodon.org/entities/Account/#acct
	acct, found := payload[acctKey].(string)
	if !found || acct == "" {
		log.Error("key %q not found in response")
		return nil
	}

	log = log.With("account", acct)

	if !h.accountOldEnough(log, payload) {
		log.With("created_at", payload["created_at"]).
			Warn("Account is not old enough, forbiding")
		r.StatusCode = http.StatusForbidden
		return nil
	}

	fakeEmail := acct
	if !strings.Contains(fakeEmail, "@") {
		// Append domain if account does not include it already.
		fakeEmail = fmt.Sprintf("%s@%s", fakeEmail, h.domain)
	}

	const fakeEmailKey = "fake_email"
	payload[fakeEmailKey] = fakeEmail

	newBody := &bytes.Buffer{}
	err = json.NewEncoder(newBody).Encode(payload)
	if err != nil {
		log.Error("couldn't encode modified payload in json", "err", err)
		return nil
	}

	log.With(fakeEmailKey, fakeEmail).Info("added fake email")

	// For some wicked reason httputil.ReverseProxy does not do this for me.
	r.Header.Set("content-length", strconv.Itoa(newBody.Len()))

	r.Body = io.NopCloser(newBody)
	return nil
}

func (h Handler) accountOldEnough(log *slog.Logger, payload map[string]any) bool {
	if h.opts.MinAge == 0 {
		return true
	}

	oldEnough, err := timestampOlderThan(payload["created_at"], h.opts.MinAge)
	if err != nil {
		log.Error("checking account age using 'created_at'", "err", err)
		// Something went wrong, default to allow.
		return true
	}

	return oldEnough
}

func timestampOlderThan(timestampRaw any, threshold time.Duration) (bool, error) {
	if threshold == 0 {
		return true, nil
	}

	createdAtStr, isStr := timestampRaw.(string)
	if !isStr {
		return false, fmt.Errorf("JSON object is not a string")
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return false, fmt.Errorf("parsing creation time: %w", err)
	}

	return time.Since(createdAt) > threshold, nil
}
