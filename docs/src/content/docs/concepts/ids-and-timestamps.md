---
title: IDs and timestamps
description: Sonyflake identifier structure, JSON string encoding, and UTC timestamp formatting.
---

## Sonyflake identifiers

Users, authentication sessions, and links use Sonyflake identifiers. Sonyflake
is a distributed ID format inspired by Twitter's Snowflake. Each identifier
combines elapsed time, a sequence, and a generator identifier.

The service uses the Sonyflake v2 defaults:

| Component | Bits | Meaning |
| --- | ---: | --- |
| Elapsed time | 39 | 10 ms units since `2025-01-01T00:00:00Z` |
| Sequence | 8 | Up to 256 IDs within one 10 ms unit per generator |
| Generator ID | 16 | Distinguishes generators that may issue IDs concurrently |

The three components occupy 63 bits, so every generated value is a positive
signed 64-bit integer. Generator coordination is managed by the service and
does not require client input.

:::note
An ID carries time and machine information. It is not a secret, random token,
or authorization mechanism. Never grant access merely because a client knows
an ID.
:::

## Why IDs are strings

JSON has one number type, and JavaScript represents ordinary numbers as IEEE
754 double-precision values. Integers are exact only through
`9,007,199,254,740,991` (`2^53 - 1`), while a positive 63-bit Sonyflake may be
as large as `9,223,372,036,854,775,807`.

Returning an ID as a JSON number could silently round it and cause a client to
request or delete the wrong resource. The API therefore encodes every ID as a
base-10 string:

```json
{
  "id": "229605206530457602",
  "user_id": "229605206530457601"
}
```

Keep IDs as strings throughout parsing, storage, comparison, routing, and UI
state. Do not convert them to JavaScript `number` values.

## Timestamp format

API timestamps are UTC strings in RFC 3339 format:

```text
2026-10-10T14:32:15Z
2026-10-10T14:32:15.123Z
2026-10-10T14:32:15.123456789Z
```

Fractional seconds are omitted when zero and otherwise use only the digits
needed, up to nanoseconds. Clients must parse RFC 3339 rather than assuming a
fixed number of fractional digits.

Timestamp precision may vary between responses. Clients must compare parsed
instants rather than timestamp text and must not rely on a particular number of
fractional digits.

Use the explicit `created_at`, `updated_at`, and `expires_at` fields for business
logic. Clients should treat the time portion encoded inside an ID as an
implementation detail.
