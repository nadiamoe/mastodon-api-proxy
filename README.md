# `mastodon-api-proxy`

A proxy to Mastodon API that adds a `fakeEmail` field to the [`Account`](https://docs.joinmastodon.org/entities/Account/) object returned by [`verify_credentials`](https://docs.joinmastodon.org/methods/accounts/#verify_credentials).

Every other response, even `Account` responses from other endpoints, are left untouched.

### Usage

Run the container. `BACKEND_URL` is expected to point to the real mastodon URL. `DOMAIN` is the email domain used to form `fakeEmail`, in the form of `{acct}@{DOMAIN}`, where `acct` is the `acct` field returned by Mastodon.

### Why?

https://github.com/requarks/wiki/discussions/7037
