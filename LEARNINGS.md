# Engineering Learnings

A running log of backend-engineering lessons surfaced while building Hookwave, written
generally enough to carry into future projects — not just fixes for this codebase.

**How to use this doc:** each entry follows the same shape — the general principle, how to
spot it, and how to fix it. When you hit a new lesson (in this project or another), add a new
entry in the same template rather than a one-off note. Group by category; add a new category
heading if something doesn't fit an existing one.

---

## 1. Layered architecture & dependency direction

### Principle
In a layered app (handler → service → repository → domain), dependencies should point
**inward only**. Each layer's public API should be expressed in types it owns — never in a
type borrowed from a layer further out. "Further out" isn't about which package imports which
in a chain — a handler *calling into* a service package is fine, that's just how you invoke
anything in Go. The violation is when an inner layer's own method **signature** is shaped by
an outer layer's type.

### Red flag
A service method that takes a `dto.*` type, or a domain type that references anything from
`service`/`api`. Ask: "could I call this function from something that isn't an HTTP handler
without needing to construct an HTTP-shaped struct first?" If no, the boundary is backwards.

### Fix
The inner layer defines its own input/output types. The outer layer translates into them at
the boundary (handler decodes JSON → dto → validates → maps to the service's own params type
→ calls service). This translation step is not boilerplate to eliminate — it *is* the
boundary.

### Example from Hookwave
`service.CreateEnvironment` originally took `*dto.CreateEnvironmentRequest` directly. Fixed by
giving `service` its own `CreateEnvironmentParams`, with the handler doing the dto → params
translation.

### Nuance: not every "outer" dependency is equally bad
`service` also references `repository.CreateEnvironmentParams` (sqlc-generated types) in its
repository-port interfaces. Same shape of "coupled to an outer layer's type" — but it's a
**deliberate, pragmatic exception**, not a law with a carve-out: sqlc's generated types are a
stable, auto-generated contract, and using them lets `*repository.Queries` satisfy the
service's interface with zero adapter code (Go interfaces are structural). The cost: driver-
specific types (`pgtype.Text`, `*json.RawMessage`) can leak into the service layer once a
table has nullable/JSON columns. Decide this one consciously per project, and watch for the
leak showing up as you add entities with more complex column types.

---

## 2. Package design: cohesion and naming

### Principle
A package should answer one question honestly: "what is this for?" If you can't answer in one
sentence, or the answer is "assorted stuff," the boundary is wrong.

### Red flag
A "utils"/"common"/"helpers" package that keeps growing. A strong tell: it needs to import
something surprising to do its job — e.g. a generic "utils" package importing a Postgres
driver package is really data-access code wearing the wrong name tag.

### Fix
Split by actual concern, not by "things that didn't have an obvious home." Name each resulting
package for what it does, so a future reader's first guess at where to find something is
usually right.

### Example from Hookwave
`internal/utils` mixed three unrelated jobs: an HTTP error envelope type (`APIError`), a
Postgres-error → domain-error translator (`MapError`, importing `pgconn` directly), and raw
HTTP response strings. Split into: an error-envelope/sentinel package, a hand-written
`errors.go` inside `internal/repository` (next to the generated files) for DB-error mapping,
and response strings moved into the one package that actually consumes them (`handlers`).

### Related: name files honestly, not by convention-cargo-culting
A file named `main.go` should be `package main` with a `func main()`. If it's not, name it for
its contents (`server.go`, `handlers.go`) — `main.go` signals something specific in Go and
using it elsewhere misleads readers and tooling expectations.

### Related: naming overlap between adjacent packages
Two packages whose names could each plausibly hold the other's content (e.g. a `db` package
that only bootstraps a connection, next to a `repository` package that holds the actual
queries) reads ambiguous to a new reader. If two packages are "the same kind of thing"
(infra bootstrap, say), group them under a shared parent (`platform/db`, `platform/logging`)
so the relationship is visible in the path, not just tribal knowledge.

---

## 3. Composition root: wiring belongs in exactly one place

### Principle
Every app has one moment where all the pieces get connected — DB pool → repository → service
→ handler. That moment ("the composition root") should live at the single outermost point of
the program (`main.go`, or one dedicated bootstrap file it calls). No inner layer should
construct another layer's concrete dependencies itself.

### Red flag
A package importing something it only needs for *constructing* another layer, never for
calling it. E.g. a `handlers` package importing `repository` solely so an `Init()` function
inside `handlers` can build a `service.XService(dbConn)` — even though every handler only ever
calls the `XService` interface, never `repository` directly.

### Fix
Give the inner package a plain constructor that accepts already-built dependencies
(`NewHandlers(envSvc EnvironmentService) *Handlers`), and do the actual construction chain in
`main.go` (or `NewServer`, if you want `main.go` to stay minimal). This shrinks what the inner
package needs to import down to only the interfaces it actually uses.

---

## 4. `context.Context` must stay immutable — never store one in a mutable struct

### Principle
The stdlib is explicit: don't store a `Context` inside a struct type and mutate it; each layer
should wrap the previous context and return a **new** value (`context.WithValue` never
mutates its input). This is what makes contexts safe to pass to goroutines and safe for
multiple unrelated pieces of middleware to layer values onto without stepping on each other.

### Red flag
- A custom type that implements `Deadline()/Done()/Err()/Value()` itself (i.e. re-implements
  the `context.Context` interface) instead of just being a `context.Context`.
- A method like `SetValue` that reassigns a struct field holding a context, rather than
  returning a new context.
- `ctx.(*SomeConcreteType)` anywhere — a blind type assertion on what should be an opaque
  `context.Context` — because it silently assumes nothing between where the value was set and
  where it's read ever wraps the context the normal way.

### Fix
Use the standard pattern: an unexported key type, a `WithX(ctx, val) context.Context` setter
that calls `context.WithValue`, and a getter `XFromContext(ctx) (T, bool)` using a comma-ok
type assertion so a missing value is a handled case, not a panic.

```go
type ctxKey int
const loggerKey ctxKey = iota

func WithLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}
func LoggerFromContext(ctx context.Context) (*Logger, bool) {
	l, ok := ctx.Value(loggerKey).(*Logger)
	return l, ok
}
```

---

## 5. Error handling discipline

### Principle
Compare errors with `errors.Is`/`errors.As`, never `==` or a raw type assertion, because errors
get wrapped (`fmt.Errorf("...: %w", err)`) as code evolves — `==` silently stops matching the
moment that happens, with no compile-time warning.

### Red flag
`err == someSentinel` sitting right next to a correct `errors.As(err, &someType)` two lines
away — a sign the inconsistency wasn't a deliberate choice, just momentum.

### Fix
`errors.Is(err, sentinel)` / `errors.As(err, &target)`, always, even when today the error
happens to reach the check unwrapped.

### Related: don't blanket-handle every error the same way
A handler that catches any service error and returns the same status/message for all of them
(a genuine 404, and a genuine 500, both becoming "400 Bad Request: X is not valid") is a real
correctness bug, not a style nit — it actively misleads API callers about what happened.
Branch on the specific sentinel errors you expect (`errors.Is(err, ErrResourceNotFound)` → 404,
etc.), same as you already do consistently in most of your handlers.

---

## 6. Multi-tenant scoping must be enforced uniformly, not per-operation

### Principle
If a tenant key (`environment_id`, `org_id`, whatever the boundary is) scopes reads, it must
scope **every** operation on that data — create implicitly does (via the FK), but read,
update, and delete all need it explicitly, or the isolation guarantee has a hole.

### Red flag
One query in an entity's CRUD set scoping by the tenant key while its siblings don't — e.g.
`GetByID(id, tenantID)` scoped correctly, `DeleteByID(id)` not. This kind of gap is easy to
introduce because the failure mode is invisible until there's a second tenant to leak across.

### Fix
Scope every query that touches tenant-owned data by the tenant key, and make the interface
signature itself require it (don't let `DeleteByID(ctx, id)` even compile without a tenant
param) — a missing parameter is much harder to accidentally ship than a missing `WHERE` clause
a reviewer has to remember to check for.

---

## 7. Naming: constructors and types

### Principle
- **Plain data types** (no behavior): name them for the entity only, no layer/role suffix —
  the package already provides that context (`domain.Environment`, not
  `domain.EnvironmentDomain` — avoid stutter).
- **Behavior types** (services, handlers, repository interfaces — anything with methods that
  *act*): keep the role suffix (`EnvironmentService`, `EnvironmentHandler`) — here the suffix
  tells you what *kind* of actor it is, which the package name alone doesn't convey once a
  package holds more than one such type.
- **Constructors**: `New()` alone only if the package has exactly one constructible type
  (`logging.New()`, `repository.New()`). If a package has several (`service.EnvironmentService`
  and `service.ApplicationService`), each constructor needs the suffix
  (`NewEnvironmentService`, `NewApplicationService`) or `New()` is ambiguous.
- A constructor's name must match what it **returns** — `NewEnvironment()` returning
  `*EnvironmentService` is a mismatch (worse if a `domain.Environment` type already exists
  elsewhere, since the name collides with a different concept entirely).

### Fix
Before naming a constructor, ask: (1) does this package have one constructible type or many?
(2) does the name I'm about to give it match the type it returns?

---

## 8. Process: verify after every rename/refactor

### Principle
A package rename or file move is only "done" when `go build ./...` and `go vet ./...` are both
clean — grep for the old import path too, since Go's compiler will only catch what's actually
reached at build time, and a stale import in a file that's otherwise fine still breaks the
whole build.

### Red flag
Renaming a package/function and only manually updating the call sites you remember, instead of
grepping for every reference to the old name.

### Fix
After any rename: `grep -rl "old/import/path" --include="*.go"`, `go build ./...`,
`go vet ./...`, `gofmt -l .` — all four, every time, before considering the rename finished.
