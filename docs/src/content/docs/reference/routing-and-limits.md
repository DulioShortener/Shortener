---
title: Routing and limits
description: API versioning, content types, body limits, rate limits, CORS, and redirect routes.
---

## Route boundaries

Versioned JSON endpoints live below:

```text
https://3dreamstudio.com.br/api/v1
```

Authentication commands use `/api/v1/auth`; account resources use
`/api/v1/users`; link resources use `/api/v1/links`. The older-looking
`/api/v1/login` path does not exist and returns `404 ROUTE_NOT_FOUND`.

Two browser-facing routes deliberately sit outside the API prefix:

- `GET /r/{code}` returns a `302` redirect for a short code.
- `GET /` and `HEAD /` return a `301` redirect to
  `https://app.3dreamstudio.com.br/`.

## Request and response bodies

Send request bodies as UTF-8 JSON with `Content-Type: application/json`. The
server rejects malformed JSON with `422 INVALID_FORM_BODY` and unsupported body
media types with `415 UNSUPPORTED_MEDIA_TYPE`.

The global request-body limit is 65,536 bytes (64 KiB). An oversized body
returns `413 REQUEST_BODY_TOO_LARGE` before an endpoint processes it.

## Account and link limits

| Limit | Value |
| --- | ---: |
| Links per account | 50 |
| URL size | 4096 UTF-8 bytes |
| Short-code length | 8 Base62 characters |
| Password length on registration | 8–128 Unicode characters |
| Username length | 2–32 ASCII characters |
| Display-name length | 2–32 Unicode characters |

## Authentication rate limit

`POST /api/v1/users` and `POST /api/v1/auth/login` share one token bucket per
client IP address on each API instance:

- sustained rate: 5 requests per second;
- burst: 5 requests; and
- idle in-memory entries expire after 3 minutes.

Responses from those two routes include `X-RateLimit-Limit` and
`X-RateLimit-Remaining`. A denied request also includes `Retry-After` and
returns `429 RATE_LIMIT_EXCEEDED`.

Because the limiter is in memory, counters are not shared across API instances
and reset when an instance restarts.

## Request IDs

Every response includes `X-Request-Id`. If the request supplies that header,
the service reuses it; otherwise it generates a 32-character value. Record the
ID when troubleshooting a failed request.

## CORS

The service allows browser requests from any origin and permits requested
preflight headers such as `Authorization` and `Content-Type`. Credentialed CORS
is disabled: browsers do not send cookies or HTTP credentials to this API.

The permissive origin policy is compatible with bearer headers but does not
make a token public. Applications remain responsible for protecting tokens.
