---
title: Dulio Shortener API
description: The complete contract for creating accounts, authenticating, and managing short links.
---

Dulio Shortener exposes a small HTTPS API for account-scoped short links. This
reference documents the complete public contract: every route, field, error,
limit, normalization rule, and response format implemented by the service.

## Base URLs

| Purpose | URL |
| --- | --- |
| Versioned JSON API | `https://3dreamstudio.com.br/api/v1` |
| Public short links | `https://3dreamstudio.com.br/r/{code}` |
| Browser application | `https://app.3dreamstudio.com.br/` |

The API does **not** use an `api.` subdomain. The documentation hostname is
separate from the service and does not change request URLs.

## Create a link

After [logging in](/guides/authentication/), send the returned bearer token
with the destination URL. A successful request returns the immutable link.

<div class="quick-request-grid">
  <section aria-labelledby="quick-request-title">
    <p id="quick-request-title" class="quick-request-label">Request</p>
    <pre tabindex="0"><code>curl --request POST \
  --url https://3dreamstudio.com.br/api/v1/links \
  --header 'Authorization: Bearer YOUR_TOKEN' \
  --header 'Content-Type: application/json' \
  --data '{"url":"https://example.com/articles/sonyflake"}'</code></pre>
  </section>
  <section aria-labelledby="quick-response-title">
    <p id="quick-response-title" class="quick-request-label">201 response</p>
    <pre tabindex="0"><code>{
  "id": "229605206530457602",
  "user_id": "229605206530457601",
  "code": "aZ09BcDe",
  "target_url": "https://example.com/articles/sonyflake",
  "short_url": "https://3dreamstudio.com.br/r/aZ09BcDe",
  "created_at": "2026-10-10T14:34:02.842Z"
}</code></pre>
  </section>
</div>

Example identifiers, tokens, and timestamps in this site are illustrative.

:::tip[Use the live reference]
Open the [interactive API reference](/api-reference/) to inspect schemas,
generate client examples, and send requests to production.
:::

## Operations

| Method | Path | Authentication | Purpose |
| --- | --- | --- | --- |
| `POST` | `/api/v1/users` | No | Create an account |
| `GET` | `/api/v1/users/me` | Bearer | Get the current account |
| `POST` | `/api/v1/auth/login` | No | Create a bearer session |
| `POST` | `/api/v1/auth/logout` | Bearer | Revoke the current session |
| `POST` | `/api/v1/links` | Bearer | Create a short link |
| `GET` | `/api/v1/links` | Bearer | List the current account's links |
| `DELETE` | `/api/v1/links/{id}` | Bearer | Delete an owned link |
| `GET` | `/r/{code}` | No | Redirect to a stored destination |
| `GET`, `HEAD` | `/` | No | Redirect to the browser application |

## Representation rules

- IDs are positive Sonyflake values returned as decimal JSON strings.
- Timestamps are UTC RFC 3339 strings with optional fractional seconds.
- Successful JSON responses use `application/json`.
- Empty successes use HTTP `204` with no response body.
- Failures use one stable [`errors` envelope](/reference/errors/).
- Authenticated operations expect `Authorization: Bearer <token>`.

Continue with [Getting started](/guides/getting-started/) for a complete account
and link lifecycle.
