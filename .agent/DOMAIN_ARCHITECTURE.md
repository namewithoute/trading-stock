# 🎯 TRADING SYSTEM - DOMAIN ARCHITECTURE

## 📋 DOMAIN OVERVIEW

Hệ thống trading được chia thành **10 domain chính** theo nguyên tắc Domain-Driven Design (DDD):

### 1. **User Domain** - Quản lý người dùng
**Entities:**
- User (ID, Email, Username, Password, Status)
- Profile (KYC information)
- Role & Permission

**Use Cases:**
- Register user
- Login/Logout
- Update profile
- KYC verification

**API Endpoints:**
```
POST   /api/v1/users/register
POST   /api/v1/users/login
GET    /api/v1/users/profile
PUT    /api/v1/users/profile
POST   /api/v1/users/kyc
```

---

### 2. **Account Domain** - Quản lý tài khoản giao dịch
**Entities:**
- TradingAccount (ID, UserID, Type, Balance, BuyingPower)
- Transaction (Deposit, Withdrawal)
- MarginInfo

**Use Cases:**
- Create trading account
- Deposit funds
- Withdraw funds
- Get account balance
- Calculate buying power

**API Endpoints:**
```
POST   /api/v1/accounts
GET    /api/v1/accounts/:id
POST   /api/v1/accounts/:id/deposit
POST   /api/v1/accounts/:id/withdraw
GET    /api/v1/accounts/:id/balance
```

---

### 3. **Order Domain** - Quản lý lệnh giao dịch
**Entities:**
- Order (ID, UserID, Symbol, Price, Quantity, Side, Type, Status)
- OrderBook (collection of orders)

**Value Objects:**
- Side (BUY, SELL)
- OrderType (MARKET, LIMIT, STOP_LOSS, STOP_LIMIT)
- OrderStatus (PENDING, FILLED, CANCELLED, REJECTED)

**Use Cases:**
- Create order (market/limit)
- Cancel order
- Modify order
- Get order status
- List user orders

**API Endpoints:**
```
POST   /api/v1/orders
GET    /api/v1/orders/:id
DELETE /api/v1/orders/:id
PUT    /api/v1/orders/:id
GET    /api/v1/orders/user/:user_id
```

---

### 4. **Portfolio Domain** - Quản lý danh mục đầu tư
**Entities:**
- Portfolio (ID, UserID, TotalValue, Cash)
- Position (Symbol, Quantity, AvgPrice, CurrentPrice, P&L)
- Holding (historical positions)

**Use Cases:**
- Get current positions
- Calculate P&L (realized/unrealized)
- Get portfolio performance
- Get asset allocation

**API Endpoints:**
```
GET    /api/v1/portfolios/:user_id
GET    /api/v1/portfolios/:user_id/positions
GET    /api/v1/portfolios/:user_id/performance
GET    /api/v1/portfolios/:user_id/pnl
```

---

### 5. **Market Data Domain** - Dữ liệu thị trường
**Entities:**
- Stock (Symbol, Name, Exchange)
- Price (Symbol, Price, Timestamp)
- Candle (OHLCV data)
- MarketDepth (Bid/Ask levels)

**Use Cases:**
- Get real-time price
- Get historical data (candles)
- Get market depth
- Subscribe to price updates (WebSocket)

**API Endpoints:**
```
GET    /api/v1/market/stocks/:symbol
GET    /api/v1/market/stocks/:symbol/price
GET    /api/v1/market/stocks/:symbol/candles
GET    /api/v1/market/stocks/:symbol/depth
WS     /ws/market/prices
```

---

### 6. **Execution Domain** - Thực thi lệnh
**Entities:**
- Execution (ID, OrderID, Price, Quantity, Timestamp)
- Fill (partial or full execution)
- Trade (matched order)

**Use Cases:**
- Execute order
- Report fill
- Match orders (if building own exchange)
- Send order to broker

**Events:**
- OrderExecuted
- OrderFilled
- OrderPartiallyFilled

---

### 7. **Risk Management Domain** - Quản lý rủi ro
**Entities:**
- RiskLimit (MaxPositionSize, MaxOrderValue)
- RiskMetric (VaR, Exposure)
- ComplianceRule

**Use Cases:**
- Pre-trade risk check
- Calculate position risk
- Check margin requirements
- Validate compliance rules

**Business Rules:**
- Maximum position size per symbol
- Maximum order value
- Pattern day trader rules
- Margin requirements

---

### 8. **Notification Domain** - Thông báo
**Entities:**
- Notification (ID, UserID, Type, Message, Status)
- Alert (PriceAlert, OrderAlert)

**Use Cases:**
- Send order status notification
- Send price alert
- Send email/SMS/push notification

**Channels:**
- Email
- SMS
- Push notification
- WebSocket

---

### 9. **Analytics Domain** - Phân tích & Báo cáo
**Entities:**
- Report (TradeHistory, P&L Report, Tax Report)
- Dashboard (Custom metrics)

**Use Cases:**
- Generate trade history report
- Calculate tax documents
- Create performance dashboard
- Export data

---

### 10. **Strategy Domain** (Optional - Algo Trading)
**Entities:**
- Strategy (ID, Name, Rules, Parameters)
- Signal (BUY/SELL signal)
- Backtest (historical performance)

**Use Cases:**
- Define trading strategy
- Backtest strategy
- Execute automated trading
- Paper trading

---

## 🏗️ CLEAN ARCHITECTURE LAYERS

```
┌─────────────────────────────────────────────────────────┐
│                    PRESENTATION LAYER                    │
│  (HTTP Handlers, WebSocket, gRPC, GraphQL)              │
│  - Receive requests                                      │
│  - Validate input                                        │
│  - Call use cases                                        │
│  - Return responses                                      │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                    USE CASE LAYER                        │
│  (Business Logic, Application Services)                 │
│  - Orchestrate business workflows                       │
│  - Coordinate between repositories                      │
│  - Enforce business rules                               │
│  - Publish domain events                                │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                    DOMAIN LAYER                          │
│  (Entities, Value Objects, Repository Interfaces)       │
│  - Define business entities                             │
│  - Define repository contracts (interfaces)             │
│  - No external dependencies                             │
│  - Pure business logic                                  │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                  INFRASTRUCTURE LAYER                    │
│  (Repository Implementations, External Services)        │
│  - PostgreSQL repositories                              │
│  - Redis caching                                         │
│  - Kafka event publishing                               │
│  - External broker APIs                                 │
└─────────────────────────────────────────────────────────┘
```

---

## 🔄 DATA FLOW EXAMPLE: Create Order

```
1. HTTP Request
   POST /api/v1/orders
   {
     "user_id": "user123",
     "symbol": "AAPL",
     "price": 150.00,
     "quantity": 10,
     "side": "BUY",
     "order_type": "LIMIT"
   }

2. Handler Layer (handler/http/order_handler.go)
   ↓ Validate request
   ↓ Call use case

3. Use Case Layer (usecase/order/create_order.go)
   ↓ Check buying power (call Account domain)
   ↓ Validate business rules (Risk domain)
   ↓ Create order entity
   ↓ Save to repository

4. Repository Layer (repository/postgres/order_repo.go)
   ↓ Save to PostgreSQL
   ↓ Return order ID

5. Use Case Layer (continued)
   ↓ Publish OrderCreated event to Kafka
   ↓ Send notification

6. Handler Layer (continued)
   ↓ Return HTTP response
   {
     "order_id": "order123",
     "status": "PENDING"
   }
```

---

## 📊 DATABASE SCHEMA DESIGN

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

### Accounts Table
```sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    account_type VARCHAR(20) NOT NULL,
    balance DECIMAL(20, 2) NOT NULL DEFAULT 0,
    buying_power DECIMAL(20, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    INDEX idx_user_id (user_id)
);
```

### Orders Table
```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    symbol VARCHAR(10) NOT NULL,
    price DECIMAL(20, 4) NOT NULL,
    quantity INT NOT NULL,
    side VARCHAR(4) NOT NULL,
    order_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    filled_quantity INT DEFAULT 0,
    avg_fill_price DECIMAL(20, 4),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_symbol (symbol),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

### Positions Table
```sql
CREATE TABLE positions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    symbol VARCHAR(10) NOT NULL,
    quantity INT NOT NULL,
    avg_price DECIMAL(20, 4) NOT NULL,
    current_price DECIMAL(20, 4),
    unrealized_pnl DECIMAL(20, 2),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(account_id, symbol),
    INDEX idx_user_id (user_id)
);
```

---

## 🔥 EVENT-DRIVEN ARCHITECTURE

### Kafka Topics

```
trading.orders.created       - Order created events
trading.orders.filled        - Order filled events
trading.orders.cancelled     - Order cancelled events
trading.positions.updated    - Position updated events
trading.accounts.deposited   - Deposit events
trading.accounts.withdrawn   - Withdrawal events
trading.market.prices        - Real-time price updates
trading.notifications        - Notification events
```

### Event Example

```json
// Topic: trading.orders.created
{
  "event_id": "evt123",
  "event_type": "OrderCreated",
  "timestamp": "2026-02-03T15:30:00Z",
  "data": {
    "order_id": "order123",
    "user_id": "user123",
    "symbol": "AAPL",
    "price": 150.00,
    "quantity": 10,
    "side": "BUY",
    "order_type": "LIMIT"
  }
}
```

---

## 🎯 IMPLEMENTATION PRIORITY

### MVP (Minimum Viable Product)
1. ✅ User Domain - Registration & Login
2. ✅ Account Domain - Basic balance tracking
3. ✅ Order Domain - Create & cancel orders
4. ✅ Market Data Domain - Real-time prices

### Phase 2
5. Portfolio Domain - Position tracking
6. Execution Domain - Order execution
7. Notification Domain - Alerts

### Phase 3
8. Risk Management Domain
9. Analytics Domain
10. Strategy Domain (if needed)

---

## 📝 CODING STANDARDS

### Repository Pattern
```go
// Domain layer defines interface
type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
}

// Infrastructure layer implements
type postgresOrderRepo struct {
    db *gorm.DB
}

func (r *postgresOrderRepo) Create(ctx context.Context, order *Order) error {
    return r.db.WithContext(ctx).Create(order).Error
}
```

### Use Case Pattern
```go
type CreateOrderUseCase struct {
    orderRepo order.Repository
    accountRepo account.Repository
    logger *zap.Logger
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, input CreateOrderInput) (*CreateOrderOutput, error) {
    // 1. Validate
    // 2. Check business rules
    // 3. Create entity
    // 4. Save to repository
    // 5. Publish event
    // 6. Return result
}
```

### Handler Pattern
```go
type OrderHandler struct {
    createOrderUC *order.CreateOrderUseCase
    logger *zap.Logger
}

func (h *OrderHandler) CreateOrder(c echo.Context) error {
    var input order.CreateOrderInput
    if err := c.Bind(&input); err != nil {
        return c.JSON(400, ErrorResponse{Error: "Invalid input"})
    }
    
    output, err := h.createOrderUC.Execute(c.Request().Context(), input)
    if err != nil {
        return c.JSON(500, ErrorResponse{Error: err.Error()})
    }
    
    return c.JSON(201, output)
}
```

---

## 🚀 NEXT STEPS

1. Review domain boundaries
2. Start implementing User domain (complete flow)
3. Add database migrations
4. Implement authentication middleware
5. Add comprehensive tests
6. Deploy to staging environment

Bạn muốn bắt đầu implement domain nào trước?
