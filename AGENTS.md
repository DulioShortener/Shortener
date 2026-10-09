# Repository Working Rules

## Engineering quality bar

- Write production-quality code for both the Go backend and the React
  frontend. Optimize for correctness, clarity, explicit behavior, security,
  accessibility, and maintainability rather than cleverness or brevity.
- Prefer small, cohesive modules with one clear reason to change. Do not create
  generic dumping grounds such as broad `utils`, `common`, or catch-all model
  files.
- Keep abstractions narrow and purposeful. Add a layer, interface, shared
  component, hook, or dependency only when it protects a real boundary or
  removes demonstrated duplication.
- Treat errors and edge states as part of the product contract. Do not leak
  infrastructure details to clients, silently discard failures, or leave
  loading, empty, error, and unauthorized states undefined.
- Review design as part of every change. Challenge code that works but weakens
  cohesion, dependency direction, testability, accessibility, or the domain
  model. Update `ARCHITECTURE.md` whenever a boundary or dependency direction
  changes.

## Backend architecture and Go

- Follow the pragmatic hexagonal architecture documented in
  `ARCHITECTURE.md`; use clean-code principles without adding ceremonial
  layers that do not protect a real boundary.
- Dependencies must point inward. Entities and services must not depend on
  Echo, SQLite, HTTP request or response types, or concrete security adapters.
- Define interfaces in the consuming package. Accept narrow interfaces where
  substitution is needed and return concrete implementations from
  constructors.
- HTTP handlers own transport concerns only: binding, validation, extracting
  request identity, invoking a use case, and formatting the response. Business
  rules belong in services, and SQL belongs in repositories.
- The `app` package is the backend composition root and the only place that
  wires concrete adapters together. Avoid package globals, hidden service
  locators, and constructors that perform unrelated work.
- Pass `context.Context` explicitly as the first parameter of operations that
  may block or perform I/O. Do not store request contexts in long-lived
  structs.
- Keep entities, request types, response types, handlers, services, and
  repositories in focused feature files. Preserve error causes with `%w` and
  translate them only at the appropriate boundary.
- Run `gofmt` on every touched Go file. Keep imports, names, resource cleanup,
  and error handling consistent with idiomatic Go; code that merely compiles
  is not considered finished.

## Frontend architecture and React

- Keep `frontend` as an independently buildable React, TypeScript, and Vite
  application. Use TypeScript strictly; do not introduce unjustified `any`,
  unsafe casts, ignored type errors, or disabled lint rules.
- Grow `frontend/src` by feature ownership rather than by file type. A feature
  should own its components, hooks, state, and tests. Put code in a shared area
  only after it is genuinely reusable, and do not create empty architectural
  directories in anticipation of future work.
- Keep dependency direction explicit: application composition may depend on
  pages or features, features may depend on shared primitives, and shared code
  must not import feature internals. Features must not reach into one another's
  private files.
- Centralize HTTP behavior behind a typed API boundary. Presentational
  components must not issue ad hoc requests, know endpoint construction, or
  interpret backend error envelopes. Keep transport DTOs distinct from view
  state when their shapes or lifecycles differ.
- Keep server state, application state, and ephemeral UI state distinct. Do
  not copy derived values into state, add global state for local concerns, or
  scatter authentication credential handling across components.
- Prefer composition and narrow, explicit props over boolean-prop
  proliferation. Components must remain pure during render; use effects only
  to synchronize with external systems. The React Compiler is enabled, so do
  not add manual memoization without measured need.
- Build accessible interfaces by default: use semantic HTML, associated
  labels, keyboard-operable controls, visible focus, meaningful status and
  error messaging, and sensible responsive behavior.
- Keep styling scoped to its owner and reserve global CSS for resets, tokens,
  and deliberate application-wide rules. Reuse established design tokens and
  interaction patterns instead of creating near-duplicates.
- Use `npm` and keep `frontend/package-lock.json` synchronized. Before adding a
  dependency, justify its bundle cost, maintenance health, security exposure,
  and overlap with React, the web platform, or existing packages.

## Tests and verification

- Add or update automated tests for every behavioral change. Cover critical
  paths and failure modes completely; a happy-path-only test suite is not
  sufficient.
- Database changes require a real temporary-file SQLite integration test using
  the embedded migrations. Never replace this with a repository mock alone.
- A frontend behavioral change requires an appropriate unit, component, or
  integration test. If no frontend test harness exists yet, the first such
  change must establish the smallest suitable harness rather than leaving the
  behavior untested.
- For backend changes, run formatting, all Go tests, the race detector, `vet`,
  and relevant container builds. For frontend changes, run `npm run lint`,
  `npm run build`, and the configured test command. Resolve warnings and type
  errors rather than suppressing them.
- Contract changes must be tested at both sides of the boundary when both are
  affected. Keep IDs as strings and timestamps as ISO 8601 strings throughout
  frontend transport types.

## Project constraints

- Keep schema changes in ordered Goose SQL files under `database/migration`.
  Never run migrations automatically from API startup.
- Maintain a single application `README.md` at the repository root; do not add
  a separate frontend or backend README.
- Files below `.agents/` are tool-owned and must remain untouched.
- Do not perform remote Git operations unless explicitly requested.
