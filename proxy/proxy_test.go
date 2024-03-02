package proxy_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/tidwall/gjson"
	"roob.re/mastodon-api-proxy/proxy"
)

func Test_Handler(t *testing.T) {
	t.Parallel()

	backend := httptest.NewServer(fakeMastodon())
	t.Cleanup(func() {
		backend.Close()
	})

	proxy, err := proxy.New(backend.URL, "test.local")
	if err != nil {
		t.Fatalf("building proxy: %v", err)
	}

	server := httptest.NewServer(proxy)
	t.Cleanup(func() {
		server.Close()
	})

	t.Run("does not touch other endpoints", func(t *testing.T) {
		t.Parallel()

		response, err := http.Get(server.URL + "/something/else")
		if err != nil {
			t.Fatalf("requestinsomething else: %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Fatalf("request returned %d", response.StatusCode)
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("reading body: %v", err)
		}

		if !bytes.Equal(body, []byte(`{"imA":"fake json document"}`)) {
			t.Fatalf("response does not match expected: %s", string(body))
		}
	})

	for _, tc := range []struct {
		name   string
		acct   string
		status int
		assert func(t *testing.T, r *http.Response)
	}{
		{
			name:   "creates email from local account",
			acct:   "foo",
			status: http.StatusOK,
			assert: func(t *testing.T, r *http.Response) {
				if r.StatusCode != http.StatusOK {
					t.Fatalf("unexpected status code %d", r.StatusCode)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("reading body: %v", err)
				}

				if fe := gjson.GetBytes(body, "fake_email").String(); fe != "foo@test.local" {
					t.Fatalf("unexpected value for fake_email %q", fe)
				}
			},
		},
		{
			name:   "respects email from remote account",
			acct:   "foo@bar.local",
			status: http.StatusOK,
			assert: func(t *testing.T, r *http.Response) {
				if r.StatusCode != http.StatusOK {
					t.Fatalf("unexpected status code %d", r.StatusCode)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("reading body: %v", err)
				}

				if fe := gjson.GetBytes(body, "fake_email").String(); fe != "foo@bar.local" {
					t.Fatalf("unexpected value for fake_email %q", fe)
				}
			},
		},
		{
			name:   "propagates non-ok status codes",
			acct:   "something",
			status: http.StatusBadRequest,
			assert: func(t *testing.T, r *http.Response) {
				if r.StatusCode != http.StatusBadRequest {
					t.Fatalf("unexpected status code %d", r.StatusCode)
				}
			},
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/accounts/verify_credentials", nil)
			if err != nil {
				t.Fatalf("creating request: %v", err)
			}

			req.Header.Add("acct", tc.acct)
			req.Header.Add("echo-status", fmt.Sprint(tc.status))

			response, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("making request: %v", err)
			}

			tc.assert(t, response)
		})
	}
}

func fakeMastodon() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/something/else", func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`{"imA":"fake json document"}`))
	})
	mux.HandleFunc("/api/v1/accounts/verify_credentials", func(rw http.ResponseWriter, r *http.Request) {
		echoStatus, err := strconv.Atoi(r.Header.Get("echo-status"))
		if err != nil {
			panic(err)
		}

		rw.WriteHeader(echoStatus)
		// Example from https://docs.joinmastodon.org/methods/accounts/#verify_credentials
		_, _ = fmt.Fprintf(rw, `{
				"id": "14715",
				"username": "trwnh",
				"acct": %q,
				"display_name": "infinite love ⴳ",
				"locked": false,
				"bot": false,
				"created_at": "2016-11-24T10:02:12.085Z",
				"note": "<p>i have approximate knowledge of many things. perpetual student. (nb/ace/they)</p><p>xmpp/email: a@trwnh.com<br /><a href=\"https://trwnh.com\" rel=\"nofollow noopener noreferrer\" target=\"_blank\"><span class=\"invisible\">https://</span><span class=\"\">trwnh.com</span><span class=\"invisible\"></span></a><br />help me live: <a href=\"https://liberapay.com/at\" rel=\"nofollow noopener noreferrer\" target=\"_blank\"><span class=\"invisible\">https://</span><span class=\"\">liberapay.com/at</span><span class=\"invisible\"></span></a> or <a href=\"https://paypal.me/trwnh\" rel=\"nofollow noopener noreferrer\" target=\"_blank\"><span class=\"invisible\">https://</span><span class=\"\">paypal.me/trwnh</span><span class=\"invisible\"></span></a></p><p>- my triggers are moths and glitter<br />- i have all notifs except mentions turned off, so please interact if you wanna be friends! i literally will not notice otherwise<br />- dm me if i did something wrong, so i can improve<br />- purest person on fedi, do not lewd in my presence<br />- #1 ami cole fan account</p><p>:fatyoshi:</p>",
				"url": "https://mastodon.social/@trwnh",
				"avatar": "https://files.mastodon.social/accounts/avatars/000/014/715/original/34aa222f4ae2e0a9.png",
				"avatar_static": "https://files.mastodon.social/accounts/avatars/000/014/715/original/34aa222f4ae2e0a9.png",
				"header": "https://files.mastodon.social/accounts/headers/000/014/715/original/5c6fc24edb3bb873.jpg",
				"header_static": "https://files.mastodon.social/accounts/headers/000/014/715/original/5c6fc24edb3bb873.jpg",
				"followers_count": 821,
				"following_count": 178,
				"statuses_count": 33120,
				"last_status_at": "2019-11-24T15:49:42.251Z"
			}`, r.Header.Get("acct"))
	})

	return mux
}
