---
title: Links and URLs
description: URL validation, duplicate detection, link limits, short codes, and redirect behavior.
---

Links are immutable account-owned mappings from an eight-character short code
to an HTTP or HTTPS destination.

## Accepted destinations

`POST /api/v1/links` accepts one required `url` field. The value must:

- be valid UTF-8;
- be no longer than 4096 UTF-8 bytes;
- be an absolute `http` or `https` URL;
- include a hostname; and
- contain no username or password information.

Surrounding whitespace is removed before the value is stored. The remaining
spelling—including hostname casing and an explicit default port—is preserved in
`target_url` and used in the redirect's `Location` header.

## Duplicate detection

The service derives a separate canonical comparison key. It:

- lowercases the scheme and hostname;
- removes port `80` from HTTP URLs and port `443` from HTTPS URLs; and
- treats an empty path as `/`.

It preserves path casing, query parameter order, query values, and fragments.
Consequently, these are duplicates for the same account:

```text
https://Example.com:443
https://example.com/
```

These are distinct because canonicalization does not reorder the query:

```text
https://example.com/?a=1&b=2
https://example.com/?b=2&a=1
```

Duplicate detection is per account. Different accounts may shorten the same
destination.

## Short codes

Codes contain exactly eight case-sensitive characters from this Base62 set:

```text
0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz
```

Codes are generated with cryptographic randomness and are unique across the
service. A code is not a Sonyflake and must not be substituted for the link
`id` in delete requests.

## Limits and ordering

- An account may own at most 50 links.
- Links cannot be edited. Delete and recreate a link to change its destination.
- `GET /api/v1/links` returns every owned link with no pagination.
- Results are sorted by `created_at` descending and then `id` descending.
- Deleting another account's link returns `404 LINK_NOT_FOUND`.

## Public redirects

`GET /r/{code}` requires no authentication. A known code returns HTTP `302`
with the stored destination in `Location`; an unknown code returns the normal
JSON `404 LINK_NOT_FOUND` envelope.
