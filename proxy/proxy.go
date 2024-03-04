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
)

type Handler struct {
	proxy  *httputil.ReverseProxy
	domain string
	logger *slog.Logger
}

const verifyCredentialsURL = "/api/v1/accounts/verify_credentials"

func New(logger *slog.Logger, backendUrl, domain string) (Handler, error) {
	u, err := url.Parse(backendUrl)
	if err != nil {
		return Handler{}, fmt.Errorf("parsing backend url: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	h := Handler{
		proxy:  proxy,
		domain: domain,
		logger: logger,
	}

	proxy.ModifyResponse = func(r *http.Response) error {
		h.logger.With("path", r.Request.URL.Path, "upstream_status", r.Status).Info("handling response")

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
	h.logger.With("path", r.URL.Path).Info("handling request")
	h.proxy.ServeHTTP(rw, r)
}

func (h Handler) addFakeEmail(r *http.Response) error {
	log := h.logger.With("path", r.Request.URL.Path, "upstream_status", r.Status)

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

	fakeEmail := fmt.Sprintf("%s@%s", acct, h.domain)
	if strings.Contains(acct, "@") {
		fakeEmail = acct
	}

	const fakeEmailKey = "fake_email"
	payload[fakeEmailKey] = fakeEmail

	newBody := &bytes.Buffer{}
	err = json.NewEncoder(newBody).Encode(payload)
	if err != nil {
		log.Error("couldn't encode modified payload in json: %v", err)
		return nil
	}

	log.With(fakeEmailKey, fakeEmail).Info("added fake email")

	// For some wicked reason httputil.ReverseProxy does not do this for me.
	r.Header.Set("content-length", strconv.Itoa(newBody.Len()))

	r.Body = io.NopCloser(newBody)
	return nil
}
