---
title: Reviews and replies
description: Rating rules, public visibility, reply threads, deletion, and pagination.
---

`/api/v1/reviews` holds ratings and conversations about Dulio Shortener. A root
review has text and a rating from 1 to 5. A reply has text only. Replies may
target a root or another reply, so threads can continue to any depth.

Each account may have one active root review. Other accounts can reply to it,
and the reviewer may also join the thread. Deleting a root review hides it and
all its replies, and allows its author to create a new root review.
Deleting a reply leaves a placeholder with `body: null`; its existing
descendants remain readable. A deleted reply
cannot receive new direct replies. Content cannot be edited.

Every review and reply includes `author_id` as a decimal string, including in
reply previews and placeholders. Responses never include user or author
objects, profile fields, or deletion flags and timestamps.

## Reading reviews

`GET /api/v1/reviews` is public and returns at most 50 active roots per page,
newest first. Without a bearer token, each item contains only the root review;
`replies` and `has_replies` are absent. A valid bearer token adds a `replies`
page to each root with at most one direct reply preview. An invalid supplied
token returns `401 UNAUTHORIZED`.

`GET /api/v1/reviews/{id}` uses the same visibility rule. Anonymous readers can
retrieve roots only; a reply ID requires authentication. The authenticated
`GET /api/v1/reviews/me` returns the caller's active root with its preview. It
returns `404 REVIEW_NOT_FOUND` if no active root exists.

## Loading a conversation

`GET /api/v1/reviews/{id}/replies` requires a bearer token and accepts either
a root ID or a reply ID. It returns at most 50 **direct** replies, oldest first.
It never nests the children of those replies. Each returned item has a
`has_replies` boolean; use that to decide whether to request the item's own
`/replies` route. The list's `has_more` is different: it reports additional
siblings after the current page.

Pages have the same shape at both routes:

```json
{
  "items": [],
  "has_more": false,
  "next_cursor": null
}
```

When `has_more` is true, pass `next_cursor` as the next request's `cursor`
query parameter. The cursor is a plain `created_at_unix_milliseconds:id`
position. Clients should treat its value as a string. Review pages sort newest
first; reply pages sort oldest first. The server fixes the page size and reply
preview depth; clients cannot request more nested content in one response.

## Posting and deletion

Creating a root requires `POST /api/v1/reviews` with a `body` and integer
`rating`. The body must contain 1–2000 Unicode characters and at least one
non-whitespace character. A second active root returns
`409 REVIEW_ALREADY_EXISTS`.

A rating outside 1–5 returns `422 INVALID_FORM_BODY` with
`errors.fields.rating: "This value must be in range of 1 and 5"`.

To reply, `POST /api/v1/reviews/{id}/replies` with a `body` and a bearer token.
The route accepts any active root or reply. A `rating` property is invalid for
replies, even when set to `null`. Deleting your own root or reply uses
`DELETE /api/v1/reviews/{id}` and returns `204`. Missing, hidden, and unowned
items return `404 REVIEW_NOT_FOUND`.
