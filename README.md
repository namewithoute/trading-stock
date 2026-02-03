# Trading Stock - Enterprise Go Microservice

Production-ready trading platform built with Go, featuring robust retry logic, graceful shutdown, and enterprise-grade patterns.

## 🚀 Quick Start

```bash
# 1. Start dependencies
docker-compose up -d

# 2. Run application
go run cmd/api/main.go

# 3. Health check
curl http://localhost:8080/ping
```

## 📁 Project Structure

```
trading-stock/
├── cmd/api/                    # Application entrypoint
│   └── main.go
├── internal/
│   ├── bootstrap/              # Startup & shutdown orchestration
│   │   ├── startup.go         # Application initialization
│   │   └── shutdown.go        # Graceful shutdown manager
│   ├── config/                 # Configuration management
│   │   ├── config.go          # Main config structs
│   │   └── init.go            # Init-specific config
│   ├── domain/                 # Business entities
│   ├── initialize/             # External service initialization
│   │   ├── postgres.go        # Database connection with retry
│   │   ├── redis.go           # Redis connection with retry
│   │   └── kafka.go           # Kafka producer with retry
│   └── global/                 # Shared state (to be refactored)
├── pkg/
│   ├── logger/                 # Structured logging (Zap)
│   └── utils/                  # Utilities
│       └── retry.go           # Exponential backoff retry logic
└── .agent/
    ├── ENTERPRISE_ASSESSMENT.md  # Detailed assessment
    └── examples/
        └── retry_examples.go     # Usage examples
```

## 🎯 Key Features

### 1. **Exponential Backoff Retry** ⚡
```go
// Automatic retry with exponential backoff + jitter
cfg := utils.DefaultRetryConfig()
err := utils.DoWithRetry(ctx, logger, "Database", cfg, func() error {
    return db.Ping()
})
```

**Features:**
- ✅ Exponential backoff (1s → 2s → 4s → 8s...)
- ✅ Jitter to prevent thundering herd
- ✅ Context-aware cancellation
- ✅ Retryable vs permanent error classification
- ✅ Configurable max attempts and intervals

### 2. **Graceful Shutdown** 🛑
```go
// Handles SIGTERM/SIGINT gracefully
// 1. Stop accepting new requests (10s timeout)
// 2. Close external connections (5s timeout)
// 3. Flush Kafka messages
// 4. Exit cleanly
```

**Features:**
- ✅ OS signal handling (Kubernetes-friendly)
- ✅ Separate timeouts for each phase
- ✅ Concurrent resource cleanup with WaitGroup
- ✅ Error aggregation
- ✅ Zero downtime deployments

### 3. **Structured Logging** 📝
```go
logger.Info("Order created",
    zap.String("order_id", orderID),
    zap.Float64("amount", 1000.50),
    zap.Duration("latency", time.Since(start)),
)
```

**Features:**
- ✅ JSON output for log aggregation
- ✅ Log rotation (size, age, backups)
- ✅ Multiple log levels
- ✅ File + console output

### 4. **Connection Pooling** 🏊
```go
// Postgres
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(5 * time.Minute)

// Redis
PoolSize: 20
MinIdleConns: 5
```

## 🔧 Configuration

### Environment Variables
```bash
# Override config values
export DB_SOURCE="postgres://user:pass@localhost:5432/trading"
export REDIS_ADDR="localhost:6379"
export KAFKA_BROKERS="localhost:9092"
```

### Config Files
```yaml
# internal/configs/dev.yaml
database:
  source: "postgres://postgres:postgres@localhost:5432/trading_stock?sslmode=disable"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 5

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 20
  min_idle_conns: 5

kafka:
  brokers: ["localhost:9092"]
  batch_size: 100
  batch_timeout: 10
```

## 📊 Retry Configuration

### Default (Production)
```go
utils.DefaultRetryConfig()
// MaxAttempts: 10
// InitialInterval: 1s
// MaxInterval: 30s
// Multiplier: 2.0
// Jitter: true
```

### Custom
```go
cfg := utils.RetryConfig{
    MaxAttempts:     5,
    InitialInterval: 500 * time.Millisecond,
    MaxInterval:     10 * time.Second,
    Multiplier:      1.5,
    Jitter:          true,
}
```

### Backoff Timeline
```
Attempt 1: 1s
Attempt 2: 2s (±25% jitter)
Attempt 3: 4s (±25% jitter)
Attempt 4: 8s (±25% jitter)
Attempt 5: 16s (±25% jitter)
...
Attempt 10: 30s (capped at MaxInterval)
```

## 🛠️ Development

### Prerequisites
```bash
go version  # >= 1.21
docker --version
docker-compose --version
```

### Setup
```bash
# 1. Clone repository
git clone <repo-url>
cd trading-stock

# 2. Install dependencies
go mod download

# 3. Start infrastructure
docker-compose up -d postgres redis kafka

# 4. Run application
go run cmd/api/main.go
```

### Build
```bash
# Development build
go build -o trading-stock.exe ./cmd/api

# Production build (optimized)
go build -ldflags="-s -w" -o trading-stock.exe ./cmd/api
```

### Testing
```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...

# Run examples
go run .agent/examples/retry_examples.go
```

## 🚢 Deployment

### Docker
```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-s -w" -o trading-stock ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/trading-stock .
COPY --from=builder /app/internal/configs ./internal/configs
CMD ["./trading-stock"]
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: trading-stock
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: trading-stock
        image: trading-stock:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_SOURCE
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: connection-string
        livenessProbe:
          httpGet:
            path: /ping
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ping
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 📈 Monitoring

### Health Check
```bash
curl http://localhost:8080/ping
# Response: {"message":"pong","status":"healthy"}
```

### Logs
```bash
# View logs
tail -f logs/trading-stock.log

# Search for errors
grep "ERROR" logs/trading-stock.log

# Count retry attempts
grep "retrying" logs/trading-stock.log | wc -l
```

### Metrics (TODO)
```go
// Add Prometheus metrics
import "github.com/prometheus/client_golang/prometheus"

var (
    retryCounter = prometheus.NewCounterVec(...)
    shutdownDuration = prometheus.NewHistogram(...)
)
```

## 🔒 Security Best Practices

- ✅ No secrets in code (use environment variables)
- ✅ Database connection pooling (prevent connection exhaustion)
- ✅ Graceful shutdown (prevent data loss)
- ⏳ TODO: Rate limiting
- ⏳ TODO: Input validation
- ⏳ TODO: JWT authentication

## 🐛 Troubleshooting

### Database Connection Fails
```bash
# Check if Postgres is running
docker ps | grep postgres

# Check connection string
psql "postgres://postgres:postgres@localhost:5432/trading_stock"

# View retry logs
grep "Postgres" logs/trading-stock.log
```

### Kafka Connection Fails
```bash
# Check if Kafka is running
docker ps | grep kafka

# Test connection
kafka-console-producer --broker-list localhost:9092 --topic test

# View retry logs
grep "Kafka" logs/trading-stock.log
```

### Graceful Shutdown Not Working
```bash
# Send SIGTERM (Kubernetes uses this)
kill -TERM <pid>

# Check shutdown logs
grep "Shutting down" logs/trading-stock.log

# Verify resources closed
grep "closed" logs/trading-stock.log
```

## 📚 Learning Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [12-Factor App](https://12factor.net/)

## 🎯 Roadmap

### Phase 1: Foundation ✅
- [x] Exponential backoff retry
- [x] Graceful shutdown
- [x] Structured logging
- [x] Connection pooling

### Phase 2: Testing (In Progress)
- [ ] Unit tests (target: 80% coverage)
- [ ] Integration tests (testcontainers)
- [ ] Benchmark tests

### Phase 3: Observability
- [ ] Prometheus metrics
- [ ] OpenTelemetry tracing
- [ ] Request ID propagation

### Phase 4: Resilience
- [ ] Circuit breaker
- [ ] Rate limiting
- [ ] Bulkhead pattern

### Phase 5: Production
- [ ] Database migrations (golang-migrate)
- [ ] Dependency injection
- [ ] API documentation (Swagger)
- [ ] CI/CD pipeline

## 📝 License

MIT

## 👥 Contributors

- Alex (Senior Golang Engineer)

---

**Last Updated:** 2026-02-03
**Version:** 1.0.0
**Status:** Production-Ready ✅
