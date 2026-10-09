# Architecture

## Purpose and shape

Dulio Shortener is a full-stack application with two independently built
runtime surfaces:

- `backend` is a Go and Echo HTTPS API using a pragmatic hexagonal
  architecture.
- `frontend` is a React, TypeScript, and Vite browser application. It is still
  a minimal scaffold; its rules below define how it should grow rather than
  claiming feature modules that do not yet exist.

The architecture protects domain behavior, dependency direction, explicit
contracts, and testability. It does not require one directory for every design
pattern or an interface in front of every concrete type.

At runtime, Cloudflare sends the application hostname to the frontend and the
base hostname to the Go origin. The browser calls the JSON API over HTTPS. A
short URL bypasses the frontend entirely:

```text
Browser application -> /api/v1/* -> HTTP adapter -> service -> repository -> SQLite
Public visitor      -> /r/:code  -> HTTP adapter -> service -> repository -> HTTP 302
```

## Repository ownership

- `backend/cmd/api` is the API executable entry point.
- `backend/internal/entity` contains persistence- and transport-independent
  domain data structures.
- `backend/internal/service` contains application use cases and the narrow
  ports they consume.
- `backend/internal/repository/sqlite` implements persistence ports with
  explicit SQL.
- `backend/internal/security` implements Argon2id, opaque-token, short-code,
  and Sonyflake adapters.
- `backend/internal/transport/http` separates request and response DTOs,
  validation, handlers, authentication middleware, and route registration.
- `backend/internal/apierrors` owns the stable external error representation
  and the mapping from expected failures to HTTP semantics.
- `backend/internal/app` is the backend composition root and the only package
  that constructs and wires concrete adapters.
- `database` owns the shared SQLite connection setup, embedded Goose
  migrations, and standalone migration executable.
- `frontend` owns the browser application, its package manifest, and its build
  configuration. Feature-specific frontend source should remain with its
  feature as the application grows.

The repository has one operational `README.md`. Architectural decisions live
here; commands and deployment instructions belong in the README rather than
being duplicated.

## Dependency rules

Backend dependencies point inward. Transport and infrastructure depend on
services and entities; services depend only on entities and consumer-owned
interfaces. Entities know nothing about HTTP, Echo, SQL, SQLite, hashing, or
token generation. Constructors return concrete implementations, while a
consumer accepts a narrow interface only where substitution or isolation is
needed.

The normal backend request flow is:

1. Echo middleware handles cross-cutting HTTP behavior.
2. A handler binds and validates a request DTO and extracts any authenticated
   principal.
3. The handler invokes one service operation with explicit inputs.
4. The service enforces business rules and calls its ports.
5. An adapter performs SQL, hashing, token generation, or ID generation.
6. The handler maps the result to a response DTO; `apierrors` maps expected
   failures to the public error envelope.

Handlers must not contain SQL or domain policy. Repositories must not decide
HTTP status codes. Services must not accept Echo contexts or transport DTOs.
`context.Context` is passed explicitly to work that may block or perform I/O;
it is never stored in a long-lived object.

The frontend follows a feature-oriented dependency direction as it expands:

```text
application composition -> pages/routes -> features -> shared primitives
                                      \-> typed API boundary
```

Shared primitives cannot import a feature, and features cannot deep-import one
another's private implementation. New directories should be introduced only
when code exists to justify them; the current small scaffold does not need an
empty enterprise-shaped folder tree.

## Frontend design

The frontend uses React 19 with Strict Mode, TypeScript, Vite, Oxlint, and the
React Compiler. Render logic remains pure and components use effects only for
synchronization with external systems. Because the compiler performs React
memoization, manual `memo`, `useMemo`, and `useCallback` require a measured or
otherwise concrete reason.

As features are implemented:

- Application entry and routing compose screens; feature modules own their
  components, hooks, state, and tests.
- Presentational components render typed data and emit user intent. They do not
  construct endpoints, call `fetch` directly, or decode the backend's error
  envelope.
- A typed API boundary owns base-path construction, bearer-token attachment,
  serialization, cancellation, and normalization of successful and failed
  responses.
- Transport DTOs mirror the API contract. Domain or view models remain
  separate where display behavior or lifecycle differs. Backend IDs stay
  strings, and timestamps cross the boundary as ISO 8601 strings.
- Server state is not duplicated into unrelated global state. Ephemeral form
  and interaction state stays close to its owner. Authentication credential
  persistence is centralized and must be evaluated for browser security before
  an implementation is selected.
- Reusable components expose small composable APIs instead of accumulating
  boolean modes. Loading, empty, error, unauthorized, and success states are
  explicit.
- Accessibility is a baseline requirement: semantic elements, associated form
  labels, keyboard access, focus management, and understandable status and
  validation messages are part of feature completion.
- Global CSS is reserved for resets, design tokens, and deliberate global
  behavior. Feature styling remains locally owned and responsive.

No client state, routing, form, request, or component library is selected by
this document. Such a dependency should be introduced only when a concrete
feature justifies it and after considering bundle size, security, maintenance,
and overlap with the platform or current packages. `npm` and
`frontend/package-lock.json` are the package-management source of truth.

## API contract

JSON endpoints live under `/api/v1`. Authentication commands are grouped under
`/api/v1/auth`, while account resources remain under `/api/v1/users`.

Users, sessions, and links use Sonyflake signed 63-bit identifiers. SQLite
stores them as `INTEGER`; API responses encode them as decimal strings to avoid
precision loss in JavaScript and other clients. All API timestamps are UTC ISO
8601 values in RFC 3339 form. SQLite stores timestamps as Unix milliseconds.

Request structs own declarative validation tags, and the HTTP validation
adapter converts validation failures into the stable `apierrors` envelope.
Echo's `HTTPStatusCoder` contract is preserved so router and middleware errors
retain statuses such as 404, 405, 413, 415, and 429 without exposing internal
error details.

API contract changes require coordinated backend request or response DTO
updates and frontend transport-type updates. UI code must not infer types from
example payloads or coerce identifiers to JavaScript numbers.

## Authentication

Passwords are stored in `users.password_hash` as encoded Argon2id values. Each
hash embeds a unique random salt and the parameters needed for verification;
plaintext passwords are never stored.

Login creates a 256-bit random bearer token. Only its SHA-256 verifier is stored
in `auth_sessions`; the raw token is returned once. Authentication middleware
hashes the presented token, performs the database lookup, checks expiration,
and places a typed principal in the request context. Handlers extract the
principal and pass the user ID explicitly into services so authorization
remains visible.

Logout deletes the current session. Deleting a user cascades to every session
and link. Browser credential storage is a frontend security decision, not a
reason to weaken the backend token contract or enable cookie credentials
implicitly.

## Links

Links are immutable. A link stores the submitted redirect target and a separate
canonical comparison key. Canonicalization lowercases the scheme and host,
removes default HTTP or HTTPS ports, and gives an empty path `/`; it preserves
path, query order, and fragment.

SQLite owns the hard invariants: `(user_id, target_url_key)` is unique, codes
are globally unique, and a trigger prevents more than 50 rows per user. The
service retries random short-code collisions five times. Ownership-sensitive
deletes include both link and user IDs and expose absence as not found.

Each concurrently running API instance requires a unique 16-bit Sonyflake
machine ID. Reusing an ID concurrently can generate identifier collisions and
is a deployment configuration error that a local generator cannot detect.

## Persistence and migrations

SQLite access uses the CGO-backed `github.com/mattn/go-sqlite3` driver. The
shared database opener enables foreign keys, WAL journal mode, a five-second
busy timeout, and a bounded four-connection pool. Native development therefore
requires a C compiler.

Ordered Goose SQL migrations live in `database/migration` and are embedded in
the standalone `database/cmd/migrate` binary. The API never creates or modifies
schema during startup. Schema changes are explicit deployment operations and
can use permissions distinct from the running application.

The Compose migration service is opt-in through its `tools` profile. Ordinary
Compose startup launches only the API and cannot apply migrations as a side
effect. Database changes require integration tests against a real temporary
SQLite file with the embedded migrations applied.

Docker confines GCC and musl development headers to discarded builder stages.
The API runtime image contains the linked application binary, Alpine's musl
runtime, and the CA certificate bundle required for trusted outbound HTTPS.

## Public routing, CORS, and TLS

The backend is the origin for `DULIO_BASE_HOSTNAME`; there is no API subdomain.
The exact root path redirects permanently to
`https://app.<DULIO_BASE_HOSTNAME>/`. `/r/:code` performs a SQLite lookup and
returns an HTTP 302 directly without loading the React application. Absolute
short URLs use HTTPS and the configured base hostname.

Echo terminates TLS on port 443 using `/data/tls/origin.pem` and
`/data/tls/origin.key`. Compose bind-mounts the host `data/` directory, so the
database and certificates survive container replacement without being copied
into an image. Cloudflare proxies the base hostname to Go and routes the `app`
hostname to the frontend deployment.

CORS currently permits every origin because browser clients authenticate with
explicit bearer tokens rather than cookies. Echo allows requested preflight
headers, while credentialed cross-origin requests remain disabled. Introducing
cookie authentication or cross-origin credentials requires replacing this
policy deliberately; `*` cannot be combined safely with credentialed CORS.

## Change and testing rules

Architectural boundaries are verified by review and by tests at the closest
meaningful layer. Services use narrow fakes for business-rule tests;
repositories and migrations use real temporary-file SQLite integration tests;
HTTP behavior uses handler or server tests. Frontend behavior should use the
smallest appropriate unit, component, or integration test and must validate
user-visible error and edge states, not only successful rendering.

The frontend currently has no test harness. The first behavioral frontend
implementation must establish a suitable minimal harness and a package script
for it. Until a concrete need exists, this document intentionally does not
mandate a testing library.

Any change to a boundary, data flow, persistence invariant, deployment
topology, or public contract must update this document in the same change.
Operational commands remain in `README.md`.
