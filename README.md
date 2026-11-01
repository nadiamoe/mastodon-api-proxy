# `mastodon-api-proxy`

> **Archived.** This project is no longer maintained.
>
> The same `fake_email` field can be added by patching Mastodon directly. Apply this to your instance:
>
> ```diff
> From a36c7a4aae5384724be366b95b7d99bef7d5a057 Mon Sep 17 00:00:00 2001
> From: Nadia Santalla <nadia@nadia.moe>
> Date: Sun, 1 Nov 2026 15:01:11 +0100
> Subject: [PATCH] owocafe: export fake_email in credentials
>
> ---
>  app/serializers/rest/credential_account_serializer.rb | 1 +
>  1 file changed, 1 insertion(+)
>
> diff --git a/app/serializers/rest/credential_account_serializer.rb b/app/serializers/rest/credential_account_serializer.rb
> index 2828f60fbf..420651dff7 100644
> --- a/app/serializers/rest/credential_account_serializer.rb
> +++ b/app/serializers/rest/credential_account_serializer.rb
> @@ -2,6 +2,7 @@
>
>  class REST::CredentialAccountSerializer < REST::AccountSerializer
>    attributes :source
> +  attribute(:fake_email) { object.local_username_and_domain }
>
>    has_one :role, serializer: REST::RoleSerializer
>
> --
> 2.56.0
> ```
>
> Wiki.js v3 now needs proper OIDC support. OAuth is not enough, so the wiki.js setup below no longer works on v3. For a working OIDC bridge to Mastodon, see [dex-config.yaml](dex-config.yaml), which relies on the fake_email patch above.
>
> Pre-archival readme follows, which assumed wiki.js v2 and did not use neither the mastodon patch nor dex.

A proxy to Mastodon API that adds a `fake_email` field to the [`Account`](https://docs.joinmastodon.org/entities/Account/) object returned by [`verify_credentials`](https://docs.joinmastodon.org/methods/accounts/#verify_credentials).

Every other response, even `Account` responses from other endpoints, are left untouched.

## Usage

Run the container. `BACKEND_URL` is expected to point to the real mastodon URL. `DOMAIN` is the email domain used to form `fake_email`, in the form of `{acct}@{DOMAIN}`, where `acct` is the `acct` field returned by Mastodon.

`MIN_AGE` allows defining a minimum _account_ age (time since the account was created, not the age of the person using the account). If defined, requests to `verify_credentials` for accounts below the minimum age will return `403 Forbidden`.

`MAX_STATUS_AGE` allows defining a maximum tolerable time for the account's last post. If defined, requests to `verify_credentials` for accounts whose last post is older than the specified value will return `403 Forbidden`.

## Why?

https://github.com/requarks/wiki/discussions/7037

### So how do I use this with wiki.js?

On wiki.js:

- Use your usual mastodon url for `Authorization Endpoint URL`, e.g. `https://your.instance/oauth/authorize`
- Use your usual mastodon url for `Token Endpoint URL`, e.g. `https://your.instance/oauth/token`
- Use the URL of this thing for `User info Endpoint URL`, e.g. `https://api-proxy.your.instance/api/v1/accounts/verify_credentials`. It must be an https URL or mastodon will attempt to redirect it to https.
- As `ID Claim`, `Display Name Claim` and `Email Claim` use `id`, `display_name` and `fake_email` respectively. `fake_email` is what this proxy adds.

On your mastodon instance:

- Add the domain where this thing listens (e.g. `api-proxy.your.instance`) to the `ALTERNATE_DOMAINS` environment variable for mastodon-web.

And you're good to go!
