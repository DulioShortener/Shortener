---
title: Data models
description: Public request and response entities, field presence, nullability, and constraints.
---

The tables below describe the JSON contract. Internal persistence fields are
listed separately because clients must never depend on them.

## User creation request

| Field | Type | Required | Nullable | Constraints |
| --- | --- | --- | --- | --- |
| `username` | string | Yes | No | 2–32 ASCII lowercase letters, digits, `_`, or `.`; no consecutive periods |
| `display_name` | string | No | Yes | 2–32 Unicode characters; no leading or trailing whitespace; not blank |
| `password` | string | Yes | No | 8–128 Unicode characters; at least one lowercase, uppercase, digit, and punctuation or symbol character |

Unknown JSON properties are currently ignored. Clients should not send them or
depend on that behavior remaining part of the contract.

## Login request

| Field | Type | Required | Nullable | Constraints |
| --- | --- | --- | --- | --- |
| `username` | string | Yes | No | Same lowercase username grammar as account creation |
| `password` | string | Yes | No | Non-empty; at most 128 Unicode characters |

## User

| Field | Type | Always present | Nullable | Meaning |
| --- | --- | --- | --- | --- |
| `id` | decimal string | Yes | No | User Sonyflake ID |
| `username` | string | Yes | No | Exact lowercase username |
| `display_name` | string | Yes | Yes | Optional public display name |
| `created_at` | RFC 3339 string | Yes | No | Account creation time |
| `updated_at` | RFC 3339 string | Yes | No | Last account update time; no update operation is currently public |

The password hash is internal and is never returned.

## Authentication session

The login response represents a newly created authentication session:

| Field | Type | Always present | Nullable | Meaning |
| --- | --- | --- | --- | --- |
| `id` | decimal string | Yes | No | Session Sonyflake ID |
| `token` | string | Yes | No | 43-character opaque bearer token returned once |
| `token_type` | string | Yes | No | Always `Bearer` |
| `user` | [User](#user) | Yes | No | Account authenticated by the session |
| `created_at` | RFC 3339 string | Yes | No | Session creation time |
| `expires_at` | RFC 3339 string | Yes | No | Time after which the token is rejected |

The session's `user_id` and stored SHA-256 token verifier are internal fields.
The raw token is not recoverable after the login response is produced.

## Link creation request

| Field | Type | Required | Nullable | Constraints |
| --- | --- | --- | --- | --- |
| `url` | string | Yes | No | Absolute HTTP(S) URL; hostname required; no user information; maximum 4096 UTF-8 bytes |

## Link

| Field | Type | Always present | Nullable | Meaning |
| --- | --- | --- | --- | --- |
| `id` | decimal string | Yes | No | Link Sonyflake ID |
| `user_id` | decimal string | Yes | No | Owning user's Sonyflake ID |
| `code` | string | Yes | No | Eight case-sensitive Base62 characters |
| `target_url` | URI string | Yes | No | Stored redirect destination |
| `short_url` | URI string | Yes | No | Absolute public URL for the short code |
| `created_at` | RFC 3339 string | Yes | No | Link creation time |

The canonical `target_url_key` used for duplicate detection is internal and is
never returned.

## Link collection

`GET /api/v1/links` returns an object with one always-present `links` array.
The array contains zero to 50 [Link](#link) objects and is never `null`.

## Error

Every JSON failure wraps one error body in an `errors` property:

| Field | Type | Always present | Meaning |
| --- | --- | --- | --- |
| `errors.code` | string | Yes | Stable machine-readable code |
| `errors.message` | string | No | Request-level explanation |
| `errors.fields` | object of strings | No | Field names mapped to validation explanations |

See [Errors](/reference/errors/) for all public codes and examples.
