# Handler Package

Package `handler` chứa tất cả HTTP handlers cho Trading Stock API.

## 📁 Structure

```
internal/handler/
├── enter.go              # HandlerGroup - Entry point
├── auth_handler.go       # Authentication handlers
├── user_handler.go       # User management handlers
├── account_handler.go    # Trading account handlers
├── order_handler.go      # Order management handlers
├── portfolio_handler.go  # Portfolio handlers
├── market_handler.go     # Market data handlers
├── trade_handler.go      # Trade history handlers
└── admin_handler.go      # Admin handlers
```

## 🎯 Usage

### 1. Import HandlerGroup

```go
import "trading-stock/internal/handler"
```

### 2. Create HandlerGroup (without services - for testing)

```go
handlers := handler.NewHandlerGroup()

// Access individual handlers
handlers.AuthHandler.Login(c)
handlers.UserHandler.GetProfile(c)
handlers.OrderHandler.CreateOrder(c)
```

### 3. Create HandlerGroup (with services - production)

```go
// Initialize services first
authService := service.NewAuthService(...)
userService := service.NewUserService(...)
// ... other services

// Create handler group with dependency injection
handlers := handler.NewHandlerGroup(
    authService,
    userService,
    accountService,
    orderService,
    portfolioService,
    marketService,
    tradeService,
    adminService,
)
```

### 4. Wire with Routers

```go
// Create routers with handlers
authRouter := auth.NewAuthRouter(handlers.AuthHandler)
userRouter := user.NewUserRouter(handlers.UserHandler)
orderRouter := order.NewOrderRouter(handlers.OrderHandler)
// ... other routers

// Setup main router
v1Router := v1.NewRouter(
    echo,
    authRouter,
    userRouter,
    accountRouter,
    orderRouter,
    portfolioRouter,
    marketRouter,
    tradeRouter,
    adminRouter,
)

mainRouter := router.NewMainRouter(echo, v1Router)
mainRouter.Setup()
```

## 📊 Handler Responsibilities

### AuthHandler (4 endpoints)
- `Register` - User registration
- `Login` - User login
- `RefreshToken` - Token refresh
- `Logout` - User logout

### UserHandler (5 endpoints)
- `GetPublicProfile` - Get public user profile
- `GetProfile` - Get current user profile
- `UpdateProfile` - Update user profile
- `VerifyEmail` - Verify email address
- `SubmitKYC` - Submit KYC documents

### AccountHandler (6 endpoints)
- `VerifyAccountExists` - Check if account exists
- `ListAccounts` - List user's trading accounts
- `CreateAccount` - Create new trading account
- `GetAccountDetail` - Get account details
- `Deposit` - Deposit money
- `Withdraw` - Withdraw money

### OrderHandler (5 endpoints)
- `CreateOrder` - Place new order
- `ListOrders` - List user's orders
- `GetOrderDetail` - Get order details
- `CancelOrder` - Cancel order
- `UpdateOrder` - Update order

### PortfolioHandler (4 endpoints)
- `GetOverview` - Get portfolio overview
- `ListPositions` - List all positions
- `GetPosition` - Get position by symbol
- `GetPerformance` - Get portfolio performance

### MarketHandler (9 endpoints)
- `GetTrendingStocks` - Get trending stocks
- `ListStocks` - List all stocks
- `GetStockDetail` - Get stock details
- `GetCurrentPrice` - Get current price
- `GetCandles` - Get candlestick data
- `GetOrderBook` - Get order book
- `GetPremiumAnalysis` - Get premium analysis (protected)
- `GetWatchlist` - Get user's watchlist (protected)
- `AddToWatchlist` - Add to watchlist (protected)

### TradeHandler (3 endpoints)
- `GetMarketTrades` - Get market trades for symbol
- `ListTrades` - List user's trades
- `GetTradeDetail` - Get trade details

### AdminHandler (4 endpoints)
- `ListUsers` - List all users
- `ApproveKYC` - Approve/reject KYC
- `ListAllOrders` - List all orders in system
- `GetSystemStats` - Get system statistics

## 🔄 Dependency Flow

```
Infrastructure (DB, Redis, Kafka)
    ↓
Repositories (Data Access Layer)
    ↓
Services (Business Logic Layer)
    ↓
Handlers (HTTP Layer) ← HandlerGroup
    ↓
Routers (Route Registration)
    ↓
Echo Server
```

## 📝 TODO Implementation Checklist

Each handler has TODO comments for implementation:

- [ ] Parse request body/params
- [ ] Validate input
- [ ] Call service layer
- [ ] Handle errors
- [ ] Return response

## 🎯 Best Practices

1. **Handlers should be thin** - Business logic belongs in services
2. **Always validate input** - Use validator package
3. **Handle errors properly** - Return appropriate HTTP status codes
4. **Use DTOs** - Don't expose internal models directly
5. **Log important events** - Use structured logging

## 🚀 Next Steps

1. Implement Service Layer (`internal/service/`)
2. Implement Repository Layer (`internal/repository/`)
3. Create DTOs (`internal/dto/`)
4. Implement JWT package (`pkg/jwt/`)
5. Wire everything in `app.go`
