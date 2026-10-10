---
title: Errors
description: Error envelope, validation behavior, status codes, and stable machine-readable codes.
---

JSON failures use a stable envelope with one nested error body:

```json
{
  "errors": {
    "code": "UNAUTHORIZED",
    "message": "Authentication is required"
  }
}
```

Validation failures use `fields` instead of, or in addition to, a request-level
message:

```json
{
  "errors": {
    "code": "INVALID_FORM_BODY",
    "fields": {
      "username": "Use 2-32 lowercase letters, digits, underscores, or periods without consecutive periods"
    }
  }
}
```

Use `errors.code` for program logic. Human-readable messages may become clearer
without changing the code. A field map is unordered and may contain more than
one invalid field.

## Domain and validation codes

| Status | Code | When it occurs |
| ---: | --- | --- |
| `400` | `INVALID_LINK_ID` | Delete path parameter is not a positive signed 63-bit decimal string |
| `401` | `INVALID_CREDENTIALS` | Login username or password is incorrect |
| `401` | `UNAUTHORIZED` | Bearer header is missing or malformed, or the session is unknown, expired, or revoked |
| `404` | `USER_NOT_FOUND` | The authenticated user no longer exists |
| `404` | `LINK_NOT_FOUND` | Link ID/code is unknown or the authenticated user does not own the link |
| `409` | `USERNAME_TAKEN` | Username already exists; comparison is case-insensitive in storage |
| `409` | `LINK_ALREADY_EXISTS` | The account already owns the canonical destination |
| `409` | `LINK_LIMIT_REACHED` | The account already owns 50 links |
| `422` | `INVALID_FORM_BODY` | JSON is malformed or a request field violates its contract |

Malformed JSON reports the synthetic field `body`:

```json
{
  "errors": {
    "code": "INVALID_FORM_BODY",
    "fields": {
      "body": "The request body must contain valid JSON"
    }
  }
}
```

## Transport codes

Router and middleware failures use the same envelope:

| Status | Code | Meaning |
| ---: | --- | --- |
| `400` | `BAD_REQUEST` | Generic malformed request outside field validation |
| `403` | `FORBIDDEN` | Request identity could not be extracted |
| `404` | `ROUTE_NOT_FOUND` | No route matches the path |
| `405` | `METHOD_NOT_ALLOWED` | The path exists but not for that HTTP method |
| `408` | `REQUEST_TIMEOUT` | Request timed out |
| `413` | `REQUEST_BODY_TOO_LARGE` | Body exceeds 64 KiB |
| `415` | `UNSUPPORTED_MEDIA_TYPE` | Body media type is unsupported |
| `429` | `RATE_LIMIT_EXCEEDED` | Account-creation/login limiter denied the request |
| `502` | `BAD_GATEWAY` | Upstream gateway failure |
| `503` | `SERVICE_UNAVAILABLE` | Service is temporarily unavailable |
| `500` | `INTERNAL_SERVER_ERROR` | Unexpected application or infrastructure failure |

Unexpected failures never expose SQL, stack traces, password hashes, token
verifiers, or other infrastructure details.

## Non-JSON responses

- `204 No Content` from logout and delete has no body.
- Successful `301` and `302` redirects communicate their destination through
  the `Location` header.
- An unsuccessful `/r/{code}` lookup still uses the JSON error envelope.
