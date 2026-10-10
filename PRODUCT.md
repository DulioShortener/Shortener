# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

The product is a Go and SQLite HTTP service with a React, TypeScript, and Vite
browser application. Its English-only API documentation is a separate static
Astro Starlight site generated from Markdown and an OpenAPI contract and hosted
on Cloudflare Pages.

## Users

The primary documentation readers are developers evaluating or integrating
with the Dulio Shortener API and maintainers verifying its public contract. The
current project is a school assignment rather than a long-lived commercial
platform.

## Product Purpose

Dulio Shortener lets people create accounts, authenticate, create and manage
short links, and redirect public visitors to saved HTTP or HTTPS destinations.
Success means that an integrator can understand and exercise every public route
without reading the service implementation.

## Positioning

The service exposes a deliberately small account-scoped link API with public
redirects, stable error envelopes, Sonyflake identifiers, and explicit limits.

## Operating Context

The Go API is served over HTTPS from the base hostname, the browser application
uses the `app` subdomain, and the documentation uses the `api-reference`
subdomain. Cloudflare fronts the public hostnames, while SQLite persists users,
authentication sessions, and links.

## Capabilities and Constraints

- JSON API operations are versioned under `/api/v1`; public redirects remain at
  `/r/{code}`.
- Authentication uses opaque bearer tokens returned by login.
- Users may own at most 50 immutable links. Duplicate canonical destinations
  are rejected per user.
- Public identifiers are positive Sonyflake values encoded as decimal strings.
- API timestamps are UTC RFC 3339 strings.
- Documentation is English-only and may provide an interactive production API
  client.
- The documentation should remain proportionate to a small, short-lived school
  project and avoid custom infrastructure without a concrete need.

## Brand Commitments

The product name is Dulio Shortener. The documentation follows established
developer-reference patterns while remaining its own restrained, readable
interface.

## Evidence on Hand

The repository contains the implemented router, request and response DTOs,
validation rules, services, SQLite migrations, integration tests, README, and
architecture record. It has no customer claims, benchmarks, or external proof
that should be invented for the documentation.

## Product Principles

- Document implemented behavior, including failure modes and hard limits.
- Keep machine-readable and human-readable contracts aligned.
- Reject invalid client input instead of silently correcting it.
- Prefer a small static documentation stack that is easy to abandon or revive.
- Treat accessibility, responsive reading, and copyable examples as baseline.

## Accessibility & Inclusion

The browser application and documentation must use semantic structure,
keyboard-operable controls, visible focus, sufficient contrast, and readable
responsive layouts.
