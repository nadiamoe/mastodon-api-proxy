# `mastodon-api-proxy`

A proxy to Mastodon API that adds a `fakeEmail` field to the [`Account`](https://docs.joinmastodon.org/entities/Account/) object returned by [`verify_credentials`](https://docs.joinmastodon.org/methods/accounts/#verify_credentials).

Every other response, even `Account` responses from other endpoints, are left untouched.

### Why?

https://github.com/requarks/wiki/discussions/7037
