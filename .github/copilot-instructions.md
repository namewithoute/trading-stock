# GitHub Copilot Instructions — trading-stock

## Architecture Overview

This is a **DDD + CQRS + Event Sourcing** monolith written in Go. Layers are strictly enforced:

```
Domain → Application → Presentation
              ↑
        Infrastructure (wired only in wire.go)
```

**`internal/app/wire.go` is the single composition root.** It is the ONLY file permitted to import `internal/infrastructure/...` packages. All other layers communicate exclusively through domain interfaces.

## Layer Responsibilities

| Layer | Location | Rule |
|-------|----------|------|
| Domain | `internal/domain/<context>/` | No imports from other layers; defines aggregates, events, value objects, repository interfaces |
| Application | `internal/application/<context>/` | Imports domain interfaces only; never infra |
| Infrastructure | `internal/infrastructure/<context>/` | Implements domain interfaces; owns DB/Kafka/Redis details |
| Presentation | `internal/presentation/` | Imports application use cases only; zero domain/infra imports |

## CQRS + Event Sourcing Pattern

Account and Order domains use **Event Sourcing**. The pattern for both:

- **Write side**: `EventSourcingService` — `Load()` replays events from Postgres → `RehydrateAccount/Order()`; `Save()` appends events to Postgres then publishes to Kafka.
- **Read side**: `ReadModelRepository` — queries pre-projected `account_read_models` / `order_read_models` tables directly.
- **CQRS UseCase** (e.g. `internal/application/account/usecase.go`): thin facade embedding `CommandHandler` + `QueryHandler`.
- **"Read your own writes"**: `CommandHandler` upserts the read model synchronously before returning — Kafka/Projector is a secondary recovery path.

Aggregate state is **never set directly** — always mutated via `Apply()`. Uncommitted events accumulate in `uncommittedEvents` until `Repository.Save()` drains them.

## Outbox Pattern

Domain events are written to `outbox_events` (table: `outbox_events`) **atomically** in the same DB transaction. `OutboxRelay` polls `ProcessedAt IS NULL` rows and publishes to Kafka. See `internal/infrastructure/outbox/model.go`.

## Background Workers (all started in `app.go`)

Each worker runs in its own goroutine under a shared `workerCancel` context:

- `AccountProjector` / `OrderProjector` — Kafka → read model upsert
- `OutboxRelay` — Postgres outbox → Kafka publish
- `MatchingConsumer` — `orders.accepted` → in-process matching engine → writes trades + outbox
- `OrderFillConsumer` — `trades.executed` → order aggregate state
- `AccountTradeConsumer` — `trades.executed` → fund settlement
- `MarketTradeConsumer` — `trades.executed` → price / candle tables

## Matching Engine

`internal/infrastructure/engine/engine.go` — in-process, price-time priority. Per-symbol `OrderBook` (heap-based). Publishes results to `tradeChannel` and `orderUpdateChannel`. Tests live at `internal/infrastructure/engine/engine_test.go`.

## HTTP Layer

- Framework: **Echo v4** (`github.com/labstack/echo/v4`)
- Route groups defined per domain in `internal/presentation/router/v1/<domain>/route.go`
- Three route tiers: `/api/v1/public` (open), `/api/v1/private` (JWT required), `/api/v1/admin` (JWT + role=admin)
- `AuthMiddleware` sets `"user_id"` and `"role"` into Echo context. `RequireRole(role)` MUST be chained **after** `AuthMiddleware`.

## Developer Workflows

```bash
# Start all dependencies (Postgres, Redis, Kafka in KRaft mode — no ZooKeeper)
docker-compose up -d

# Run the application (reads internal/configs/dev.yaml via Viper)
go run cmd/api/main.go

# Run matching engine tests (only test suite currently present)
go test ./internal/infrastructure/engine/...

# Build binary
go build -o bin/trading-stock.exe ./cmd/api
```

Config is loaded from `internal/configs/dev.yaml`. Key values: app port `8081`, Kafka broker `localhost:9092`, JWT access TTL `15m` / refresh `168h`.

## Key Conventions

- **`domain.Repositories`** holds only interface fields. Kafka-dependent repositories (`account.Repository`, `order.Repository`) are **NOT** in this struct — they are injected directly into `NewUsecases()` from `wire.go`.
- Each domain command/query fully typed: see `internal/application/account/commands.go` and `queries.go`.
- Don't add infra imports to `application/` or `presentation/` packages — the compiler will not catch this, but it violates the architecture.
- New domains follow the pattern: `domain/<name>/` → `infrastructure/<name>/` → `application/<name>/usecase.go` (CQRS facade) → `presentation/handler/<name>/` → `presentation/router/v1/<name>/route.go` → registered in `wire.go` and `application/usecases.go`.
- Module name: `trading-stock` (see `go.mod`).
