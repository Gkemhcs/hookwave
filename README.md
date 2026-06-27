# Hookwave

A self-hostable outbound webhook delivery service written in Go. Producers emit events; Hookwave reliably delivers them to registered HTTP endpoints with retries, dead-lettering, HMAC signing, and idempotency.

Inspired by [Svix](https://svix.com). Built from scratch to learn production-grade backend engineering.

---

## Tech Stack


| Concern         | Choice                                                     |
| --------------- | ---------------------------------------------------------- |
| Language        | Go 1.22                                                    |
| Database        | PostgreSQL (system of record + job queue)                  |
| Queue mechanism | `SELECT FOR UPDATE SKIP LOCKED` + lease/visibility timeout |
| Migrations      | `golang-migrate`                                           |
| HTTP            | Go stdlib `net/http`                                       |
| Auth            | API keys (Milestone 2+)                                    |


---

## Architecture

```
Producer
    │
    ▼
POST /messages
    │
    ▼
[Transaction]
  INSERT messages
  INSERT deliveries (status=pending) × N endpoints
    │
    ▼
[Dispatcher loop]  ←─────────────────────────┐
  SKIP LOCKED claim                            │
    │                                          │
    ├──► Worker → POST endpoint A              │
    │       success → status=delivered         │
    │                                          │
    └──► Worker → POST endpoint B              │
            failure → status=pending           │
                      next_attempt_at += backoff┘
                      attempts++
                      (exhausted → dead_lettered)
```

---

## Project Layout

```
hookwave/
  go.work                    ← Go workspace (server + sdk)
  server/
    cmd/server/main.go       ← binary entry point
    internal/
      api/                   ← HTTP handlers, middleware, router
      domain/                ← pure Go domain types
      repository/            ← all SQL lives here
      worker/                ← dispatcher + worker pool
      config/                ← env var loading
      db/                    ← connection pool setup
    migrations/              ← SQL migration files
    go.mod
  sdk/
    go.mod                   ← client SDK (later)
```

---

## Roadmap

### v1 (current) — Phases 0–2 + packaging


| Milestone                      | What                                                              | Status         | SDE 1      | SDE 2      | Senior     |
| ------------------------------ | ----------------------------------------------------------------- | -------------- | ---------- | ---------- | ---------- |
| **M1 — Foundation**            | Server boots, DB connects, migrations run, `/health` responds     | ✅ Done         | 8–12h      | 4–6h       | 2–3h       |
| **M2 — Core CRUD API**         | Create/get users, environments, applications, endpoints           | 🔲 Not started | 12–16h     | 6–8h       | 3–5h       |
| **M3 — Ingest**                | `POST /messages` — transactional outbox, fan-out to delivery rows | 🔲 Not started | 8–12h      | 4–6h       | 2–3h       |
| **M4 — Delivery Engine**       | Dispatcher + worker pool, HTTP delivery to endpoints              | 🔲 Not started | 16–24h     | 8–12h      | 4–6h       |
| **M5 — Reliability**           | Retries + backoff + jitter, dead-lettering, graceful shutdown     | 🔲 Not started | 10–16h     | 6–8h       | 3–5h       |
| **M6 — Signing + Idempotency** | HMAC-SHA256 signing, idempotency keys                             | 🔲 Not started | 6–10h      | 4–6h       | 2–3h       |
| **M7 — Packaging**             | Structured logs, Dockerfile, docker-compose                       | 🔲 Not started | 6–8h       | 3–5h       | 2–3h       |
| **Total**                      |                                                                   |                | **66–98h** | **35–51h** | **18–28h** |


> Estimates are for focused coding time without AI assistance.
> M4 (Delivery Engine) and M5 (Reliability) are where most engineers get stuck — goroutine lifecycle, channel coordination, and graceful shutdown are non-trivial to get right.
>
>
> | Level  | Rough calendar time (part-time evenings) |
> | ------ | ---------------------------------------- |
> | SDE 1  | 5–8 weeks                                |
> | SDE 2  | 3–4 weeks                                |
> | Senior | 1.5–2 weeks                              |
>

### v2 (later hardening)

- Auth — API keys scoped per application
- Per-endpoint concurrency limits + circuit breakers
- SSRF protection on outbound requests
- Full observability — OpenTelemetry, Prometheus, pprof
- Delivery attempt history + replay
- Redis-backed queue (pluggable interface, benchmarked vs Postgres)
- Horizontal scale-out + pgbouncer
- Consumer-facing portal

---

## Data Model

```
users
  └── environments
        └── applications
              ├── endpoints   (URL, signing secret, event type filters)
              └── messages    (payload, event type, idempotency key)
                    └── deliveries  (status, attempts, next_attempt_at, locked_until)
```

---

## Getting Started

### Prerequisites

- Go 1.22+
- Docker + Docker Compose

### Run locally

```bash
# coming in M1
docker-compose up
```

### Run migrations

```bash
# coming in M1
```

---

## Delivery Semantics

- **At-least-once delivery** — exactly-once is impossible across a network. Hookwave retries on failure.
- **Idempotency keys** — producers can attach an idempotency key; consumers use the `Hookwave-Delivery-Id` header to dedup.
- **HMAC signing** — every delivery is signed with the endpoint's secret. Consumers verify the `Hookwave-Signature` header.

---

## Milestone Progress Log

### M1 — Foundation

- [ ] Directory structure scaffolded
- [ ] Config loading (env vars)
- [ ] DB connection pool
- [ ] Migration runner on startup
- [ ] HTTP server + graceful shutdown
- [ ] `GET /health`
- [ ] docker-compose (app + Postgres)

### M2 — Core CRUD API

- [ ] Domain structs
- [ ] Repository layer
- [ ] Handlers (create + get per entity)
- [ ] Request validation
- [ ] Consistent error responses

### M3 — Ingest

- [ ] Messages + deliveries repository
- [ ] `POST /messages` handler
- [ ] Transactional outbox (atomic insert)
- [ ] Event type matching + fan-out
- [ ] Returns 202

### M4 — Delivery Engine

- [ ] Bounded worker pool
- [ ] Dispatcher (SKIP LOCKED claim loop)
- [ ] HTTP delivery
- [ ] Success/failure status updates

### M5 — Reliability

- [ ] Exponential backoff + jitter
- [ ] Dead lettering
- [ ] Lease expiry + reclaim
- [ ] Graceful shutdown drain

### M6 — Signing + Idempotency

- [ ] HMAC-SHA256 signing
- [ ] Timestamp header
- [ ] Idempotency key dedup

### M7 — Packaging

- [ ] Structured logging (slog)
- [ ] Dockerfile
- [ ] docker-compose (final)