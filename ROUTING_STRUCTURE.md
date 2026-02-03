# Trading Stock - Routing Structure Analysis

## 📊 ROUTING ARCHITECTURE OVERVIEW

### **Hierarchical Structure**

```
┌─────────────────────────────────────────────────────────────────┐
│                         APP LAYER                               │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  app.go                                                   │  │
│  │  - Initializes HandlerGroup                              │  │
│  │  - Creates MainRouter with HandlerGroup                  │  │
│  │  - Calls mainRouter.Setup()                              │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      MAIN ROUTER                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  router/router.go                                         │  │
│  │  - Receives: HandlerGroup                                │  │
│  │  - Decides API Version (v1, v2, v3...)                   │  │
│  │  - Creates version routers                               │  │
│  │  - Delegates to version-specific setup                   │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      V1 ROUTER                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  router/v1/router.go                                      │  │
│  │  - Receives: HandlerGroup                                │  │
│  │  - Creates domain routers from handlers                  │  │
│  │  - Organizes routes into groups:                         │  │
│  │    • /api/v1/public    (no auth)                        │  │
│  │    • /api/v1/private   (auth required)                  │  │
│  │    • /api/v1/admin     (auth + admin role)              │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                   DOMAIN ROUTERS                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  router/v1/auth/route.go                                  │  │
│  │  router/v1/user/route.go                                  │  │
│  │  router/v1/account/route.go                               │  │
│  │  router/v1/order/route.go                                 │  │
│  │  router/v1/portfolio/route.go                             │  │
│  │  router/v1/market/route.go                                │  │
│  │  router/v1/trade/route.go                                 │  │
│  │  router/v1/admin/route.go                                 │  │
│  │                                                            │  │
│  │  Each domain router:                                      │  │
│  │  - Receives specific handler                             │  │
│  │  - Implements RegisterPublicRoutes(g *echo.Group)        │  │
│  │  - Implements RegisterRoutes(g *echo.Group)              │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔄 DEPENDENCY FLOW

```
Database (PostgreSQL)
    ↓
Repositories (Data Access)
    ↓
Services (Business Logic)
    ↓
Handlers (HTTP Layer)
    ↓
HandlerGroup (Groups all handlers)
    ↓
MainRouter (Version management)
    ↓
V1Router (Route organization)
    ↓
Domain Routers (Specific endpoints)
    ↓
HTTP Endpoints
```

---

## 📝 CODE STRUCTURE

### **1. App Layer** (`internal/app/app.go`)

```go
func (a *App) wireDependencies() error {
    // Initialize repositories
    a.Repositories = repository.NewRepositories(a.DB)
    
    // Initialize services
    a.Services = service.NewServices(
        a.Repositories,
        a.Redis,
        a.Kafka,
        a.Logger,
    )
    
    // Initialize handlers
    a.Handlers = handler.NewHandlerGroup(a.Services)
    
    // ✅ SIMPLIFIED: Only pass HandlerGroup
    mainRouter := router.NewMainRouter(a.Echo, a.Handlers)
    mainRouter.Setup()
    
    return nil
}
```

**Benefits:**
- ✅ Clean and simple
- ✅ Only 3 lines for routing setup
- ✅ No need to create individual routers

---

### **2. Main Router** (`internal/router/router.go`)

```go
type MainRouter struct {
    echo     *echo.Echo
    handlers *handler.HandlerGroup
}

func NewMainRouter(e *echo.Echo, handlers *handler.HandlerGroup) *MainRouter {
    return &MainRouter{
        echo:     e,
        handlers: handlers,
    }
}

func (m *MainRouter) Setup() {
    // ✅ DECISION POINT: Choose API version here
    v1Router := v1.NewV1Router(m.echo, m.handlers)
    v1Router.Setup()
    
    // Future: Add v2
    // v2Router := v2.NewV2Router(m.echo, m.handlers)
    // v2Router.Setup()
}
```

**Responsibilities:**
- ✅ Manage API versioning (v1, v2, v3...)
- ✅ Create version-specific routers
- ✅ Delegate setup to version routers

---

### **3. V1 Router** (`internal/router/v1/router.go`)

```go
type Router struct {
    echo     *echo.Echo
    handlers *handler.HandlerGroup
}

func NewV1Router(e *echo.Echo, handlers *handler.HandlerGroup) *Router {
    return &Router{
        echo:     e,
        handlers: handlers,
    }
}

func (r *Router) Setup() {
    // ✅ Create domain routers from handlers
    authRouter := auth.NewAuthRouter(r.handlers.AuthHandler)
    userRouter := user.NewUserRouter(r.handlers.UserHandler)
    accountRouter := account.NewAccountRouter(r.handlers.AccountHandler)
    orderRouter := order.NewOrderRouter(r.handlers.OrderHandler)
    portfolioRouter := portfolio.NewPortfolioRouter(r.handlers.PortfolioHandler)
    marketRouter := market.NewMarketRouter(r.handlers.MarketHandler)
    tradeRouter := trade.NewTradeRouter(r.handlers.TradeHandler)
    adminRouter := admin.NewAdminRouter(r.handlers.AdminHandler)
    
    // API v1 group
    v1 := r.echo.Group("/api/v1")
    
    // Public routes (no auth)
    public := v1.Group("/public")
    authRouter.RegisterPublicRoutes(public)
    userRouter.RegisterPublicRoutes(public)
    // ...
    
    // Protected routes (auth required)
    private := v1.Group("/private")
    private.Use(middleware.AuthMiddleware())
    authRouter.RegisterRoutes(private)
    userRouter.RegisterRoutes(private)
    // ...
    
    // Admin routes (auth + admin role)
    admin := v1.Group("/admin")
    admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
    adminRouter.RegisterRoutes(admin)
}
```

**Responsibilities:**
- ✅ Create domain routers from HandlerGroup
- ✅ Organize routes into public/private/admin groups
- ✅ Apply middleware at group level

---

### **4. Domain Routers** (`internal/router/v1/*/route.go`)

**Example: Auth Router**

```go
type AuthRouter struct {
    handler *handler.AuthHandler
}

func NewAuthRouter(handler *handler.AuthHandler) *AuthRouter {
    return &AuthRouter{handler: handler}
}

// Public routes - no auth required
func (r *AuthRouter) RegisterPublicRoutes(g *echo.Group) {
    auth := g.Group("/auth")
    auth.POST("/register", r.handler.Register)
    auth.POST("/login", r.handler.Login)
    auth.POST("/refresh", r.handler.RefreshToken)
}

// Protected routes - auth required
func (r *AuthRouter) RegisterRoutes(g *echo.Group) {
    auth := g.Group("/auth")
    auth.POST("/logout", r.handler.Logout)
}
```

**Responsibilities:**
- ✅ Define specific endpoints for domain
- ✅ Separate public and protected routes
- ✅ Map routes to handler methods

---

## 🌐 COMPLETE ROUTE MAP

### **Public Routes** (No Authentication)

```
GET  /health                                    → Health check

POST /api/v1/public/auth/register              → Register new user
POST /api/v1/public/auth/login                 → Login
POST /api/v1/public/auth/refresh               → Refresh token

GET  /api/v1/public/users/:id/public           → Get public profile

GET  /api/v1/public/accounts/verify/:number    → Verify account exists

GET  /api/v1/public/market/stocks              → List stocks
GET  /api/v1/public/market/stocks/:symbol      → Get stock detail
GET  /api/v1/public/market/stocks/:symbol/price        → Current price
GET  /api/v1/public/market/stocks/:symbol/candles     → Candle data
GET  /api/v1/public/market/stocks/:symbol/orderbook   → Order book
GET  /api/v1/public/market/trending            → Trending stocks

GET  /api/v1/public/trades/market              → Market trades
```

### **Protected Routes** (Authentication Required)

```
POST /api/v1/private/auth/logout               → Logout

GET  /api/v1/private/users/me                  → Get own profile
PUT  /api/v1/private/users/me                  → Update profile
POST /api/v1/private/users/me/verify-email     → Verify email
POST /api/v1/private/users/me/kyc              → Submit KYC

GET  /api/v1/private/accounts                  → List accounts
POST /api/v1/private/accounts                  → Create account
GET  /api/v1/private/accounts/:id              → Account detail
POST /api/v1/private/accounts/:id/deposit      → Deposit
POST /api/v1/private/accounts/:id/withdraw     → Withdraw

POST /api/v1/private/orders                    → Create order
GET  /api/v1/private/orders                    → List orders
GET  /api/v1/private/orders/:id                → Order detail
DELETE /api/v1/private/orders/:id              → Cancel order
PUT  /api/v1/private/orders/:id                → Update order

GET  /api/v1/private/portfolio                 → Portfolio overview
GET  /api/v1/private/portfolio/positions       → List positions
GET  /api/v1/private/portfolio/positions/:symbol → Position detail
GET  /api/v1/private/portfolio/performance     → Performance metrics

GET  /api/v1/private/market/analysis/:symbol   → Premium analysis
GET  /api/v1/private/market/watchlist          → Get watchlist
POST /api/v1/private/market/watchlist          → Add to watchlist

GET  /api/v1/private/trades                    → User trades
GET  /api/v1/private/trades/:id                → Trade detail
```

### **Admin Routes** (Authentication + Admin Role)

```
GET  /api/v1/admin/users                       → List all users
POST /api/v1/admin/users/:id/kyc/approve       → Approve KYC
GET  /api/v1/admin/orders                      → List all orders
GET  /api/v1/admin/stats                       → System statistics
```

---

## ✅ BENEFITS OF NEW STRUCTURE

### **1. Separation of Concerns**

| Layer | Responsibility |
|-------|---------------|
| **App** | Dependency injection only |
| **MainRouter** | API version management |
| **V1Router** | Route organization & middleware |
| **Domain Routers** | Specific endpoint mapping |

### **2. Easy to Add New Versions**

```go
// In MainRouter.Setup()
func (m *MainRouter) Setup() {
    // v1
    v1Router := v1.NewV1Router(m.echo, m.handlers)
    v1Router.Setup()
    
    // ✅ Add v2 easily
    v2Router := v2.NewV2Router(m.echo, m.handlers)
    v2Router.Setup()
}
```

### **3. Simplified App Initialization**

**Before:**
```go
// ❌ 20+ lines of router creation
authRouter := auth.NewAuthRouter(...)
userRouter := user.NewUserRouter(...)
// ... 8 more routers
v1Router := v1.NewV1Router(echo, authRouter, userRouter, ...)
mainRouter := router.NewMainRouter(echo, v1Router)
```

**After:**
```go
// ✅ 2 lines only!
mainRouter := router.NewMainRouter(a.Echo, a.Handlers)
mainRouter.Setup()
```

### **4. Testability**

```go
// Easy to test V1Router
func TestV1Router(t *testing.T) {
    e := echo.New()
    handlers := &handler.HandlerGroup{
        AuthHandler: mockAuthHandler,
        // ...
    }
    
    router := NewV1Router(e, handlers)
    router.Setup()
    
    // Test routes...
}
```

---

## 🎯 KEY DESIGN DECISIONS

### **1. HandlerGroup as Single Dependency**
- ✅ All handlers grouped in one struct
- ✅ Easy to pass around
- ✅ Single source of truth

### **2. Version Decision in MainRouter**
- ✅ Clear separation: MainRouter = version management
- ✅ V1Router = route organization
- ✅ Easy to add v2, v3 without changing app.go

### **3. Domain Routers Created Inside V1Router**
- ✅ V1Router controls its own structure
- ✅ App layer doesn't need to know about domain routers
- ✅ Encapsulation

### **4. Three-Level Route Groups**
- ✅ `/api/v1/public` - No auth
- ✅ `/api/v1/private` - Auth required
- ✅ `/api/v1/admin` - Auth + admin role
- ✅ Clear security boundaries

---

## 🚀 FUTURE EXTENSIBILITY

### **Adding V2 API**

```go
// 1. Create v2 package
internal/router/v2/
├── router.go
├── auth/route.go
├── user/route.go
└── ...

// 2. Implement V2Router
func NewV2Router(e *echo.Echo, handlers *handler.HandlerGroup) *Router {
    // Different structure, different routes
}

// 3. Add to MainRouter
func (m *MainRouter) Setup() {
    v1Router := v1.NewV1Router(m.echo, m.handlers)
    v1Router.Setup()
    
    v2Router := v2.NewV2Router(m.echo, m.handlers)  // ✅ Add here
    v2Router.Setup()
}
```

### **Adding New Domain**

```go
// 1. Create handler
type NotificationHandler struct {}

// 2. Add to HandlerGroup
type HandlerGroup struct {
    // ...
    NotificationHandler *NotificationHandler  // ✅ Add here
}

// 3. Create domain router
internal/router/v1/notification/route.go

// 4. Register in V1Router.Setup()
notificationRouter := notification.NewNotificationRouter(r.handlers.NotificationHandler)
notificationRouter.RegisterRoutes(private)
```

---

## 📊 SUMMARY

**Current Structure:**
```
App → MainRouter → V1Router → Domain Routers → Handlers
```

**Key Points:**
- ✅ **App** only knows about `HandlerGroup` and `MainRouter`
- ✅ **MainRouter** decides API versions (v1, v2, v3...)
- ✅ **V1Router** creates domain routers and organizes routes
- ✅ **Domain Routers** map specific endpoints to handlers
- ✅ Clean, scalable, testable, maintainable

**Perfect for:**
- ✅ Microservices migration
- ✅ API versioning
- ✅ Team collaboration
- ✅ Long-term maintenance
