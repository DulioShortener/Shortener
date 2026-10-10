---
title: Authentication
description: Bearer-token creation, use, expiration, and revocation.
---

Protected operations use opaque bearer sessions. The API does not use cookies,
refresh tokens, API keys, or implicit browser credentials.

## Send a token

Add the login response's `token` to the standard authorization header:

```http
Authorization: Bearer YOUR_TOKEN
```

The scheme is case-insensitive, but the token value is case-sensitive. The
header must contain exactly the scheme and one non-empty token.

## Session creation

`POST /api/v1/auth/login` creates a new session after verifying the exact
lowercase username and password. Each successful login creates an independent
session; logging in does not revoke older sessions.

The token is 32 cryptographically random bytes encoded as 43 unpadded base64url
characters. It is returned only in the login response and cannot be retrieved
again later.

:::caution
Anyone holding a token can act as that session. Do not put it in URLs, logs,
source control, screenshots, or client-side analytics.
:::

## Expiration

The login response includes `created_at` and `expires_at`. Treat `expires_at` as
authoritative. The current default lifetime is 24 hours, but clients must not
assume a fixed lifetime.

An expired session returns `401 UNAUTHORIZED`. The service also removes the
expired session when it is presented.

## Logout

`POST /api/v1/auth/logout` deletes the current session only. It does not revoke
other tokens belonging to the same user and it has no request body.

## Authentication failures

| Situation | Status | Code |
| --- | --- | --- |
| Incorrect login username or password | `401` | `INVALID_CREDENTIALS` |
| Missing or malformed bearer header | `401` | `UNAUTHORIZED` |
| Unknown, expired, or revoked token | `401` | `UNAUTHORIZED` |

Login intentionally returns the same `INVALID_CREDENTIALS` response for an
unknown username and an incorrect password.

## Interactive reference

In the [interactive API reference](/api-reference/), select the authentication
control and paste the raw token. Requests are sent directly from your browser to
`https://3dreamstudio.com.br`; the documentation site never proxies them.
