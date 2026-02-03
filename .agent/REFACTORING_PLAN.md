# 🏗️ REFACTORING PLAN - TRADING STOCK SYSTEM

## 📊 CURRENT STATE ANALYSIS

### ✅ Good Practices
- Separation of concerns with bootstrap, config, domain packages
- Graceful shutdown implementation with timeout management
- Domain folders structure prepared (user, order, portfolio, etc.)
- Clean configuration using Viper

### ❌ Issues to Fix

#### 1. **GLOBAL VARIABLES ANTI-PATTERN** (CRITICAL)
**Location:** `internal/global/section.go`
```go
var (
    Logger *zap.Logger
    Config *config.Config
    DB     *gorm.DB
    Redis  *redis.Client
    Kafka  *kafka.Writer
)
```

**Problems:**
- ❌ Impossible to test (cannot mock dependencies)
- ❌ Hidden dependencies (functions don't declare what they need)
- ❌ Race conditions in concurrent access
- ❌ Tight coupling to global state

**Solution:** ✅ COMPLETED - Created `App` container with dependency injection

---

#### 2. **INCONSISTENT DOMAIN STRUCTURE**
**Current:**
```
internal/domain/
├── order.go          ❌ File at root level
├── order_book.go     ❌ File at root level
├── order/            ✅ Folder but empty
├── user/             ✅ Folder but empty
└── portfolio/        ✅ Folder but empty
```

**Target:**
```
internal/domain/
├── order/
│   ├── entity.go          # Order struct
│   ├── repository.go      # Interface only
│   └── value_objects.go   # Side, OrderType, Status enums
├── user/
│   ├── entity.go
│   └── repository.go
└── portfolio/
    ├── entity.go
    └── repository.go
```

---

#### 3. **MISSING CLEAN ARCHITECTURE LAYERS**
- ❌ `handler/` empty - No HTTP handlers
- ❌ `repository/` empty - No data access implementations
- ❌ `usecase/` empty - No business logic layer

---

#### 4. **CONFIG FOLDER NAMING CONFUSION**
```
internal/
├── config/     ← Package for loading config
└── configs/    ← Folder with YAML files (confusing!)
```

**Solution:** Rename `configs/` → `config/` at root level

---

## 🎯 REFACTORING PHASES

### **PHASE 1: ELIMINATE GLOBAL VARIABLES** ✅ COMPLETED

#### Changes Made:
1. ✅ Created `internal/app/app.go` - Application container with DI
2. ✅ Created `internal/app/infrastructure.go` - Infrastructure initialization
3. ✅ Updated `cmd/api/main.go` - Use App container instead of global vars

#### Benefits:
- ✅ Testable code (can inject mock dependencies)
- ✅ Explicit dependencies (clear what each component needs)
- ✅ Thread-safe (no shared global state)
- ✅ Easier to understand and maintain

---

### **PHASE 2: RESTRUCTURE DOMAIN LAYER** (NEXT STEP)

#### Step 2.1: Move Domain Entities to Proper Folders

**Order Domain:**
```bash
# Move files
mv internal/domain/order.go → internal/domain/order/entity.go
mv internal/domain/order_state.go → internal/domain/order/value_objects.go
mv internal/domain/order_book.go → internal/domain/order/order_book.go
```

**Create repository interface:**
```go
// internal/domain/order/repository.go
package order

import "context"

type Repository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
    Update(ctx context.Context, order *Order) error
    Cancel(ctx context.Context, id string) error
    ListByUserID(ctx context.Context, userID string) ([]*Order, error)
}
```

#### Step 2.2: Create Other Domain Entities

**User Domain:**
```go
// internal/domain/user/entity.go
package user

import "time"

type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Email     string    `json:"email" gorm:"uniqueIndex"`
    Username  string    `json:"username" gorm:"uniqueIndex"`
    Password  string    `json:"-"` // Never expose in JSON
    Status    Status    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Status string

const (
    StatusActive   Status = "ACTIVE"
    StatusInactive Status = "INACTIVE"
    StatusSuspended Status = "SUSPENDED"
)
```

```go
// internal/domain/user/repository.go
package user

import "context"

type Repository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}
```

**Account Domain:**
```go
// internal/domain/account/entity.go
package account

import "time"

type Account struct {
    ID            string    `json:"id" gorm:"primaryKey"`
    UserID        string    `json:"user_id" gorm:"index"`
    AccountType   Type      `json:"account_type"`
    Balance       float64   `json:"balance"`
    BuyingPower   float64   `json:"buying_power"`
    Currency      string    `json:"currency"`
    Status        Status    `json:"status"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

type Type string

const (
    TypeCash   Type = "CASH"
    TypeMargin Type = "MARGIN"
)

type Status string

const (
    StatusActive   Status = "ACTIVE"
    StatusFrozen   Status = "FROZEN"
    StatusClosed   Status = "CLOSED"
)
```

---

### **PHASE 3: IMPLEMENT REPOSITORY LAYER**

#### Step 3.1: Create PostgreSQL Repositories

```go
// internal/repository/postgres/user_repo.go
package postgres

import (
    "context"
    "fmt"
    "trading-stock/internal/domain/user"
    
    "gorm.io/gorm"
)

type userRepository struct {
    db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) user.Repository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *user.User) error {
    if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
    var u user.User
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
    var u user.User
    if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
        return nil, fmt.Errorf("failed to get user by email: %w", err)
    }
    return &u, nil
}

func (r *userRepository) Update(ctx context.Context, u *user.User) error {
    if err := r.db.WithContext(ctx).Save(u).Error; err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }
    return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
    if err := r.db.WithContext(ctx).Delete(&user.User{}, "id = ?", id).Error; err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }
    return nil
}
```

#### Step 3.2: Create Order Repository

```go
// internal/repository/postgres/order_repo.go
package postgres

import (
    "context"
    "fmt"
    "trading-stock/internal/domain/order"
    
    "gorm.io/gorm"
)

type orderRepository struct {
    db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) order.Repository {
    return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, o *order.Order) error {
    if err := r.db.WithContext(ctx).Create(o).Error; err != nil {
        return fmt.Errorf("failed to create order: %w", err)
    }
    return nil
}

func (r *orderRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
    var o order.Order
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&o).Error; err != nil {
        return nil, fmt.Errorf("failed to get order: %w", err)
    }
    return &o, nil
}

func (r *orderRepository) Update(ctx context.Context, o *order.Order) error {
    if err := r.db.WithContext(ctx).Save(o).Error; err != nil {
        return fmt.Errorf("failed to update order: %w", err)
    }
    return nil
}

func (r *orderRepository) Cancel(ctx context.Context, id string) error {
    if err := r.db.WithContext(ctx).Model(&order.Order{}).
        Where("id = ?", id).
        Update("status", "CANCELLED").Error; err != nil {
        return fmt.Errorf("failed to cancel order: %w", err)
    }
    return nil
}

func (r *orderRepository) ListByUserID(ctx context.Context, userID string) ([]*order.Order, error) {
    var orders []*order.Order
    if err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Order("created_at DESC").
        Find(&orders).Error; err != nil {
        return nil, fmt.Errorf("failed to list orders: %w", err)
    }
    return orders, nil
}
```

---

### **PHASE 4: IMPLEMENT USE CASE LAYER**

#### Step 4.1: Create User Use Cases

```go
// internal/usecase/user/register.go
package user

import (
    "context"
    "fmt"
    "time"
    
    "trading-stock/internal/domain/user"
    
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,min=3,max=50"`
    Password string `json:"password" validate:"required,min=8"`
}

type RegisterOutput struct {
    UserID string `json:"user_id"`
}

type RegisterUseCase struct {
    userRepo user.Repository
}

func NewRegisterUseCase(userRepo user.Repository) *RegisterUseCase {
    return &RegisterUseCase{userRepo: userRepo}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
    // 1. Check if email already exists
    existingUser, err := uc.userRepo.GetByEmail(ctx, input.Email)
    if err == nil && existingUser != nil {
        return nil, fmt.Errorf("email already registered")
    }

    // 2. Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    // 3. Create user entity
    newUser := &user.User{
        ID:        uuid.New().String(),
        Email:     input.Email,
        Username:  input.Username,
        Password:  string(hashedPassword),
        Status:    user.StatusActive,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // 4. Save to database
    if err := uc.userRepo.Create(ctx, newUser); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return &RegisterOutput{UserID: newUser.ID}, nil
}
```

#### Step 4.2: Create Order Use Cases

```go
// internal/usecase/order/create_order.go
package order

import (
    "context"
    "fmt"
    "time"
    
    "trading-stock/internal/domain/order"
    
    "github.com/google/uuid"
    "go.uber.org/zap"
)

type CreateOrderInput struct {
    UserID    string  `json:"user_id" validate:"required"`
    Symbol    string  `json:"symbol" validate:"required"`
    Price     float64 `json:"price" validate:"required,gt=0"`
    Quantity  int     `json:"quantity" validate:"required,gt=0"`
    Side      string  `json:"side" validate:"required,oneof=BUY SELL"`
    OrderType string  `json:"order_type" validate:"required,oneof=LIMIT MARKET"`
}

type CreateOrderOutput struct {
    OrderID string `json:"order_id"`
}

type CreateOrderUseCase struct {
    orderRepo order.Repository
    logger    *zap.Logger
}

func NewCreateOrderUseCase(orderRepo order.Repository, logger *zap.Logger) *CreateOrderUseCase {
    return &CreateOrderUseCase{
        orderRepo: orderRepo,
        logger:    logger,
    }
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, input CreateOrderInput) (*CreateOrderOutput, error) {
    // 1. Validate business rules
    if err := uc.validateOrder(input); err != nil {
        return nil, fmt.Errorf("order validation failed: %w", err)
    }

    // 2. Create order entity
    newOrder := &order.Order{
        ID:        uuid.New().String(),
        UserID:    input.UserID,
        Symbol:    input.Symbol,
        Price:     input.Price,
        Quantity:  input.Quantity,
        Side:      order.Side(input.Side),
        OrderType: input.OrderType,
        Status:    "PENDING",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // 3. Save to database
    if err := uc.orderRepo.Create(ctx, newOrder); err != nil {
        uc.logger.Error("Failed to create order", zap.Error(err))
        return nil, fmt.Errorf("failed to create order: %w", err)
    }

    // 4. Publish order created event (to Kafka)
    // TODO: Implement event publishing

    uc.logger.Info("Order created successfully", 
        zap.String("order_id", newOrder.ID),
        zap.String("user_id", input.UserID),
    )

    return &CreateOrderOutput{OrderID: newOrder.ID}, nil
}

func (uc *CreateOrderUseCase) validateOrder(input CreateOrderInput) error {
    // Add business validation logic here
    // e.g., check buying power, market hours, etc.
    return nil
}
```

---

### **PHASE 5: IMPLEMENT HTTP HANDLERS**

#### Step 5.1: Create User Handler

```go
// internal/handler/http/user_handler.go
package http

import (
    "net/http"
    
    "trading-stock/internal/usecase/user"
    
    "github.com/labstack/echo/v4"
    "go.uber.org/zap"
)

type UserHandler struct {
    registerUseCase *user.RegisterUseCase
    logger          *zap.Logger
}

func NewUserHandler(registerUseCase *user.RegisterUseCase, logger *zap.Logger) *UserHandler {
    return &UserHandler{
        registerUseCase: registerUseCase,
        logger:          logger,
    }
}

// Register handles user registration
// POST /api/v1/users/register
func (h *UserHandler) Register(c echo.Context) error {
    var input user.RegisterInput
    
    // Bind request body
    if err := c.Bind(&input); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request body",
        })
    }

    // Validate input
    if err := c.Validate(input); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": err.Error(),
        })
    }

    // Execute use case
    output, err := h.registerUseCase.Execute(c.Request().Context(), input)
    if err != nil {
        h.logger.Error("Registration failed", zap.Error(err))
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Registration failed",
        })
    }

    return c.JSON(http.StatusCreated, output)
}
```

#### Step 5.2: Create Router

```go
// internal/handler/http/router.go
package http

import (
    "github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, userHandler *UserHandler, orderHandler *OrderHandler) {
    // API v1 group
    v1 := e.Group("/api/v1")

    // User routes
    users := v1.Group("/users")
    users.POST("/register", userHandler.Register)
    users.POST("/login", userHandler.Login)

    // Order routes (protected)
    orders := v1.Group("/orders")
    // orders.Use(middleware.JWT()) // Add JWT middleware
    orders.POST("", orderHandler.CreateOrder)
    orders.GET("/:id", orderHandler.GetOrder)
    orders.DELETE("/:id", orderHandler.CancelOrder)
    orders.GET("/user/:user_id", orderHandler.ListUserOrders)
}
```

---

### **PHASE 6: UPDATE APP WIRING**

Update `internal/app/app.go` to wire all dependencies:

```go
func (a *App) wireDependencies() error {
    // 1. Initialize repositories
    userRepo := postgres.NewUserRepository(a.DB)
    orderRepo := postgres.NewOrderRepository(a.DB)

    // 2. Initialize use cases
    registerUseCase := user.NewRegisterUseCase(userRepo)
    createOrderUseCase := order.NewCreateOrderUseCase(orderRepo, a.Logger)

    // 3. Initialize handlers
    userHandler := http.NewUserHandler(registerUseCase, a.Logger)
    orderHandler := http.NewOrderHandler(createOrderUseCase, a.Logger)

    // 4. Register routes
    http.RegisterRoutes(a.Echo, userHandler, orderHandler)

    return nil
}
```

---

## 📝 MIGRATION CHECKLIST

### Phase 1: Foundation ✅
- [x] Create App container (`internal/app/app.go`)
- [x] Extract infrastructure initialization
- [x] Update main.go to use DI

### Phase 2: Domain Layer
- [ ] Move `order.go` → `domain/order/entity.go`
- [ ] Move `order_state.go` → `domain/order/value_objects.go`
- [ ] Create `domain/order/repository.go` interface
- [ ] Create `domain/user/entity.go`
- [ ] Create `domain/user/repository.go`
- [ ] Create `domain/account/entity.go`
- [ ] Create `domain/account/repository.go`

### Phase 3: Repository Layer
- [ ] Create `repository/postgres/user_repo.go`
- [ ] Create `repository/postgres/order_repo.go`
- [ ] Create `repository/postgres/account_repo.go`
- [ ] Create `repository/redis/cache_repo.go`

### Phase 4: Use Case Layer
- [ ] Create `usecase/user/register.go`
- [ ] Create `usecase/user/login.go`
- [ ] Create `usecase/order/create_order.go`
- [ ] Create `usecase/order/cancel_order.go`

### Phase 5: Handler Layer
- [ ] Create `handler/http/user_handler.go`
- [ ] Create `handler/http/order_handler.go`
- [ ] Create `handler/http/router.go`
- [ ] Add middleware (auth, logging, rate limit)

### Phase 6: Cleanup
- [ ] Remove `internal/global/` package
- [ ] Remove `internal/bootstrap/` package
- [ ] Rename `internal/configs/` → `config/`
- [ ] Update all imports

---

## 🎯 BENEFITS AFTER REFACTORING

### Before (Global Variables):
```go
// ❌ Hidden dependencies, hard to test
func CreateOrder(order *Order) error {
    global.DB.Create(order)  // Where does DB come from?
    global.Kafka.WriteMessages(...)
    global.Logger.Info("Order created")
}
```

### After (Dependency Injection):
```go
// ✅ Explicit dependencies, easy to test
type CreateOrderUseCase struct {
    orderRepo order.Repository
    kafka     *kafka.Writer
    logger    *zap.Logger
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, input CreateOrderInput) error {
    // Clear what this function needs!
}
```

### Testing Becomes Easy:
```go
func TestCreateOrder(t *testing.T) {
    // Mock dependencies
    mockRepo := &MockOrderRepository{}
    mockKafka := &MockKafkaWriter{}
    mockLogger := zap.NewNop()

    // Inject mocks
    useCase := NewCreateOrderUseCase(mockRepo, mockKafka, mockLogger)

    // Test!
    result, err := useCase.Execute(context.Background(), input)
    assert.NoError(t, err)
}
```

---

## 🚀 NEXT STEPS

1. **Review this refactoring plan**
2. **Start Phase 2**: Restructure domain layer
3. **Implement one complete flow**: User Registration (end-to-end)
4. **Add tests** for each layer
5. **Continue with other domains**

Would you like me to start implementing Phase 2 (Domain Layer restructuring)?
