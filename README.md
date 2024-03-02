# `mastodon-api-proxy`

A proxy to Mastodon API that adds a `fakeEmail` field to the [`Account`](https://docs.joinmastodon.org/entities/Account/) object returned by [`verify_credentials`](https://docs.joinmastodon.org/methods/accounts/#verify_credentials).

Every other response, even `Account` responses from other endpoints, are left untouched.

## Usage

Run the container. `BACKEND_URL` is expected to point to the real mastodon URL. `DOMAIN` is the email domain used to form `fakeEmail`, in the form of `{acct}@{DOMAIN}`, where `acct` is the `acct` field returned by Mastodon.

## Why?

https://github.com/requarks/wiki/discussions/7037

### So how I use this with wiki.js?

On wiki.js:

- Use your usual mastodon url for `Authorization Endpoint URL`, e.g. `https://your.instance/oauth/authorize`
- Use your usual mastodon url for `Token Endpoint URL`, e.g. `https://owo.cafe/oauth/token`
- Use the URL of this thing for `User info Endpoint URL`, e.g. `https://api-proxy.your.instance/api/v1/accounts/verify_credentials`
- As `ID Claim`, `Display Name Claim` and `Email Claim` use `id`, `display_name` and `fakeEmail` respectively. `fakeEmail` is what this proxy adds.

On your mastodon instance:
- Add the domain where this thing listens (e.g. `api-proxy.your.instance`) to the `ALTERNATE_DOMAINS` environment variable for mastodon-web.

And you're good to go!
