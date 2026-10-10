---
title: Getting started
description: Create an account, log in, create a short link, and clean up the session.
---

This walkthrough exercises the complete API lifecycle. All request bodies are
JSON and every example uses the production base URL.

## 1. Create an account

```bash
curl --request POST \
  --url https://3dreamstudio.com.br/api/v1/users \
  --header 'Content-Type: application/json' \
  --data '{
    "username": "alicia.dev",
    "display_name": "Alicia",
    "password": "StrongPassword1!"
  }'
```

A successful request returns `201 Created` and the public user entity:

```json
{
  "id": "229605206530457601",
  "username": "alicia.dev",
  "display_name": "Alicia",
  "created_at": "2026-10-10T14:32:15.123456789Z",
  "updated_at": "2026-10-10T14:32:15.123456789Z"
}
```

Usernames must already be lowercase. The server rejects `Alicia` rather than
silently changing it to `alicia`. A provided display name must likewise arrive
without surrounding whitespace.

## 2. Log in

```bash
curl --request POST \
  --url https://3dreamstudio.com.br/api/v1/auth/login \
  --header 'Content-Type: application/json' \
  --data '{
    "username": "alicia.dev",
    "password": "StrongPassword1!"
  }'
```

The response contains the authentication-session ID, an opaque token, the
current user, and the session lifetime:

```json
{
  "id": "229605206530457603",
  "token": "4vQ0mVwC8N3Jk9S1xT6pL2aH7fR5yU0iE8bG3dK6zXQ",
  "token_type": "Bearer",
  "user": {
    "id": "229605206530457601",
    "username": "alicia.dev",
    "display_name": "Alicia",
    "created_at": "2026-10-10T14:32:15.123456789Z",
    "updated_at": "2026-10-10T14:32:15.123456789Z"
  },
  "created_at": "2026-10-10T14:33:01.004Z",
  "expires_at": "2026-10-11T14:33:01.004Z"
}
```

Keep the token secret. It is returned only when the session is created.

## 3. Create a link

```bash
curl --request POST \
  --url https://3dreamstudio.com.br/api/v1/links \
  --header 'Authorization: Bearer YOUR_TOKEN' \
  --header 'Content-Type: application/json' \
  --data '{"url":"https://example.com/articles/sonyflake?source=dulio"}'
```

The `short_url` is ready to share. Visiting it returns an HTTP `302` to the
stored `target_url`.

## 4. List links

```bash
curl --request GET \
  --url https://3dreamstudio.com.br/api/v1/links \
  --header 'Authorization: Bearer YOUR_TOKEN'
```

The response is `{ "links": [] }` when the account has no links. Otherwise,
links are ordered newest first. There is no pagination because the account
limit is 50 links.

## 5. Delete a link

Pass the link's decimal string `id`, not its eight-character `code`:

```bash
curl --request DELETE \
  --url https://3dreamstudio.com.br/api/v1/links/229605206530457602 \
  --header 'Authorization: Bearer YOUR_TOKEN'
```

Success returns `204 No Content`. Deleting an unknown link or a link owned by a
different account returns the same `404 LINK_NOT_FOUND` response.

## 6. Log out

```bash
curl --request POST \
  --url https://3dreamstudio.com.br/api/v1/auth/logout \
  --header 'Authorization: Bearer YOUR_TOKEN'
```

Logout revokes only the current session and returns `204 No Content`. Reusing
that token returns `401 UNAUTHORIZED`.
