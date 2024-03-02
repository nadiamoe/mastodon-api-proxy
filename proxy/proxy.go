package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

type Handler struct {
	proxy  *httputil.ReverseProxy
	domain string
}

const verifyCredentialsURL = "/api/v1/accounts/verify_credentials"

func New(backendUrl string, domain string) (Handler, error) {
	u, err := url.Parse(backendUrl)
	if err != nil {
		return Handler{}, fmt.Errorf("parsing backend url: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	h := Handler{
		proxy:  proxy,
		domain: domain,
	}

	proxy.ModifyResponse = func(r *http.Response) error {
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
	log.Println(r.URL.String())
	h.proxy.ServeHTTP(rw, r)
}

func (h Handler) addFakeEmail(r *http.Response) error {
	if r.StatusCode != http.StatusOK {
		return nil
	}

	curBody := r.Body
	defer curBody.Close()

	var payload map[string]any
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return fmt.Errorf("decoding upstream body: err")
	}

	const acctKey = "acct"
	// https://docs.joinmastodon.org/entities/Account/#username
	acct, found := payload[acctKey].(string)
	if !found || acct == "" {
		return fmt.Errorf("%q not found in response with status %d", acctKey, r.StatusCode)
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
		return fmt.Errorf("setting fake email in body: %w", err)
	}

	log.Printf("Added %q:%q", fakeEmailKey, fakeEmail)

	// For some wicked reason httputil.ReverseProxy does not do this for me.
	r.Header.Set("content-length", strconv.Itoa(newBody.Len()))

	r.Body = io.NopCloser(newBody)
	return nil
}
