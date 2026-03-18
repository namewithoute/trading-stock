---
name: maintain-component
description: "Maintain, refactor, or add features to a bounded context following DDD + CQRS + Clean Code. USE WHEN: user says 'maintain', 'refactor', 'add feature to', 'fix', 'update' a domain/component. Enforces layer separation, aggregate rules, event sourcing patterns, and clean code principles."
argument-hint: "Name the bounded context or component to maintain (e.g. 'account', 'order', 'market')"
---

# Maintain Component — DDD + Clean Code

Workflow for maintaining any bounded context in this trading-stock monolith. Every change MUST respect the DDD layer boundaries, CQRS split, Event Sourcing invariants, and clean code principles.

## When to Use

- Adding a new feature/field/endpoint to an existing domain
- Refactoring or fixing bugs in a bounded context
- Creating a new bounded context from scratch
- Any structural change to domain, application, infrastructure, or presentation code

## Step 0 — Identify the Target Context

Determine which bounded context is affected. Read these files to understand current state:

| Layer | Files to Read |
|-------|---------------|
| Domain | `internal/domain/<context>/` — all files |
| Application | `internal/application/<context>/` — all files |
| Infrastructure | `internal/infrastructure/<context>/` — all files |
| Presentation | `internal/presentation/handler/<context>/` and `internal/presentation/router/v1/<context>/` |
| Wiring | `internal/app/wire.go`, `internal/domain/repositories.go`, `internal/application/usecases.go` |

## Step 1 — Domain Layer (`internal/domain/<context>/`)

All changes start here. The domain layer has **ZERO imports** from other layers.

### File Structure

| File | Purpose |
|------|---------|
| `aggregate.go` | Aggregate Root (Event-Sourced) OR `entity.go` (CRUD domains) |
| `events.go` | Domain events: interface `DomainEvent` + concrete event structs |
| `behaviors.go` | Command methods on the aggregate — validate invariants THEN emit events |
| `value_objects.go` | Value Objects, enums, domain errors (`var ErrXxx = errors.New(...)`) |
| `ports.go` or `repository.go` | Outbound port interfaces (`Repository`, `ReadModelRepository`) |
| `read_model.go` | Query-optimised denormalised struct (CQRS read side) |

### Rules

1. **State mutation only via `Apply()`** — never set aggregate fields directly
2. **Behaviors validate before emitting** — check ALL guard conditions BEFORE calling `apply()`. If any guard fails, NO event is emitted
3. **Domain errors are sentinel values** — `var ErrXxx = errors.New("...")` in `value_objects.go`
4. **Value Objects have `IsValid()` methods** — self-validating types
5. **No infrastructure concerns** — no DB tags, no JSON tags on aggregates (JSON tags only on events and read models)
6. **Interfaces define WHAT, not HOW** — `Repository` interface says `Load/Save`, never mentions Postgres/Kafka

### Event-Sourced Aggregate Checklist

- [ ] `RehydrateXxx(events []DomainEvent) *XxxAggregate` — replays history
- [ ] `apply(event DomainEvent, isNew bool)` — switch on event type, mutate state, increment `Version`
- [ ] `UncommittedEvents() []DomainEvent` / `ClearUncommittedEvents()`
- [ ] `ToReadModel() *XxxReadModel` — converts current state to read model
- [ ] Each new event type has `GetEventType()`, `GetAggregateID()`, `GetOccurredAt()`

### Non-Event-Sourced Entity Checklist

- [ ] Plain struct with fields, no event machinery
- [ ] Repository interface with CRUD methods (`Create`, `GetByID`, `Update`, `Delete`, `List`)
- [ ] Domain logic in methods on the entity, not in the application layer

## Step 2 — Application Layer (`internal/application/<context>/`)

Imports **domain interfaces only**. Never imports `internal/infrastructure/...`.

### File Structure (CQRS domains)

| File | Purpose |
|------|---------|
| `usecase.go` | Thin CQRS facade: `UseCase interface` embedding `CommandHandler` + `QueryHandler` |
| `commands.go` | Command DTOs — immutable structs, no methods, no logic |
| `command_handlers.go` | `CommandHandler` interface + implementation: Load → Domain behavior → Save |
| `queries.go` | Query DTOs — immutable structs |
| `query_handlers.go` | `QueryHandler` interface + implementation: SELECT from read model |

### File Structure (simple CRUD domains)

| File | Purpose |
|------|---------|
| `usecase.go` | `UseCase` interface + implementation with injected repository interfaces |

### Rules

1. **Command pattern**: Load aggregate → call domain behavior → save → return read model
2. **"Read your own writes"**: CommandHandler returns `agg.ToReadModel()` after save (synchronous)
3. **Query side never replays events** — reads directly from `ReadModelRepository`
4. **No business logic here** — delegate ALL invariant checks to domain behaviors
5. **Each command/query is a typed struct** — never use `map[string]interface{}`

### Adding a New Command

```
1. Add command struct in commands.go
2. Add method to CommandHandler interface in command_handlers.go
3. Implement: Load → domain.Behavior() → saveAndProject()
4. If new event type needed → go back to Step 1
```

### Adding a New Query

```
1. Add query struct in queries.go
2. Add method to QueryHandler interface in query_handlers.go
3. Implement: call readRepo method → return result
4. If new read model field needed → update domain read_model.go + infra projector
```

## Step 3 — Infrastructure Layer (`internal/infrastructure/<context>/`)

Implements domain interfaces. Owns ALL technical details (Postgres, Kafka, Redis).

### File Structure

| File | Purpose |
|------|---------|
| `model.go` | DB model structs with GORM/pgx tags — map to/from domain entities |
| `store.go` or `repository.go` | Implements `domain.Repository` / `ReadModelRepository` |
| `event_store.go` | Postgres event store (Event-Sourced domains only) |
| `event_sourcing_service.go` | Implements `domain.Repository` via ES: Load = replay, Save = append + publish |
| `projector.go` | Kafka consumer → upsert read model (Event-Sourced domains only) |
| `*_consumer.go` | Other Kafka consumers for cross-domain integration |

### Rules

1. **DB models are separate from domain entities** — map with `ToEntity()` / `FromEntity()` methods
2. **Event serialization** — events serialize to JSON for Postgres `event_data` column
3. **Outbox pattern** — domain events written to `outbox_events` in same DB transaction
4. **Projector is idempotent** — checks `Version` to skip duplicate/out-of-order events
5. **Never import application layer** — infrastructure implements domain ports only

## Step 4 — Presentation Layer (`internal/presentation/`)

Imports **application use cases only**. Zero domain or infrastructure imports.

### File Structure

| File | Purpose |
|------|---------|
| `handler/<context>/handler.go` | HTTP handlers — parse request → call use case → format response |
| `handler/<context>/dto.go` | Request/Response DTOs with JSON tags + `ToXxxResponse()` mappers |
| `router/v1/<context>/route.go` | Route registration — `RegisterPublicRoutes()` + `RegisterRoutes()` |

### Rules

1. **Handlers are thin** — parse input, call use case, return JSON. No business logic
2. **DTOs map read models to responses** — `ToAccountResponse(rm *AccountReadModel)` in `dto.go`
3. **Request validation via struct tags** — `validate:"required"` on request DTOs
4. **Route tiers**: public (no auth), protected (JWT), admin (JWT + role check)
5. **Router implements `SubRouter` interface** — `RegisterPublicRoutes(g)` + `RegisterRoutes(g)`
6. **UserID from context** — `c.Get("user_id").(string)` set by `AuthMiddleware`

## Step 5 — Wiring (`internal/app/wire.go`)

The ONLY file that imports `internal/infrastructure/...`.

### When Adding a New Domain

1. Register new repository implementations in `infrastructure.NewRepositories()` or build them in `wire.go`
2. Add to `domain.Repositories` struct if it's a simple repo (not Kafka-dependent)
3. Inject into `application.NewUsecases()` — add parameter if ES/Kafka-dependent
4. Create handler in `presentation/handler/<context>/`
5. Create router in `presentation/router/v1/<context>/`
6. Register handler in `handler.NewHandlerGroup()`
7. Register router in `router/v1/enter.go` → `Setup()`

## Clean Code Principles (Always Enforce)

### Naming

- **Packages**: short, lowercase, singular noun (`account`, `order`, `market`)
- **Interfaces**: describe behavior, not implementation (`Repository`, not `PostgresRepository`)
- **Errors**: `Err` prefix + PascalCase (`ErrInsufficientBalance`)
- **Events**: past tense (`MoneyDepositedEvent`, `OrderAcceptedEvent`)
- **Commands**: imperative (`CreateAccountCommand`, `DepositCommand`)
- **Queries**: descriptive (`GetAccountQuery`, `ListAccountsQuery`)

### Structure

- **One concept per file** — don't mix aggregates with value objects
- **Horizontal separator comments** — `// ── Section ───` for visual grouping
- **Comment blocks before major sections** — explain WHY, not WHAT
- **No god files** — split by responsibility: behaviors, events, value objects

### Dependencies

- **Dependency flows inward only**: Presentation → Application → Domain ← Infrastructure
- **Constructor injection** — `NewXxx(deps...) Xxx` pattern everywhere
- **Return interfaces, accept interfaces** — except for concrete config structs
- **Logger is passed via constructor** — never global

### Error Handling

- **Domain errors are sentinel values** — defined in domain, checked with `errors.Is()`
- **Wrap errors with context** — `fmt.Errorf("loading account %s: %w", id, err)`
- **Don't swallow errors** — always return or log
- **Validate at boundaries** — domain behaviors validate invariants, handlers validate request format

## Verification Checklist

After every change, verify:

- [ ] **No cross-layer imports** — domain imports nothing; application imports only domain; presentation imports only application
- [ ] **wire.go is the only file importing infrastructure** outside of infrastructure itself
- [ ] **Aggregate state changed only via Apply()** (Event-Sourced domains)
- [ ] **New events registered** in `apply()` switch and infrastructure serializer
- [ ] **Read model updated** if new fields added
- [ ] **Projector updated** if event structure changed
- [ ] **DTOs updated** in presentation layer for new fields
- [ ] **Routes registered** for new endpoints
- [ ] **Code compiles** — `go build ./...`
- [ ] **Tests pass** — `go test ./...`
