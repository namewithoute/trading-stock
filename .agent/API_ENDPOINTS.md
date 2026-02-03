# 📋 TRADING STOCK API ENDPOINTS

## 🎯 CURRENT STATUS

### **✅ Implemented:**
- `GET /health` - Health check endpoint

### **📝 To Be Implemented:**
- All business endpoints below

---

## 🏗️ API STRUCTURE

```
/health                          # Health check
/api/v1/
├── auth/                        # Authentication & Authorization
├── users/                       # User Management
├── accounts/                    # Trading Accounts
├── orders/                      # Order Management
├── portfolio/                   # Portfolio & Positions
├── market/                      # Market Data
├── trades/                      # Trade History
└── admin/                       # Admin Operations
```

---

## 📚 DETAILED ENDPOINTS

### **1. HEALTH & STATUS**

#### `GET /health`
**Status:** ✅ **IMPLEMENTED**

**Description:** Check application health

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-02-03T16:38:54+07:00",
  "service": "trading-stock-api"
}
```

---

### **2. AUTHENTICATION & AUTHORIZATION** (`/api/v1/auth`)

#### `POST /api/v1/auth/register`
**Status:** ❌ Not implemented

**Description:** Register new user

**Request:**
```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}
```

**Response:**
```json
{
  "user_id": "uuid-123",
  "email": "user@example.com",
  "username": "johndoe",
  "status": "ACTIVE",
  "kyc_status": "PENDING",
  "created_at": "2026-02-03T16:38:54+07:00"
}
```

---

#### `POST /api/v1/auth/login`
**Status:** ❌ Not implemented

**Description:** User login

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "user_id": "uuid-123",
    "email": "user@example.com",
    "username": "johndoe"
  }
}
```

---

#### `POST /api/v1/auth/refresh`
**Status:** ❌ Not implemented

**Description:** Refresh access token

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

---

#### `POST /api/v1/auth/logout`
**Status:** ❌ Not implemented

**Description:** User logout (invalidate tokens)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "message": "Logged out successfully"
}
```

---

### **3. USER MANAGEMENT** (`/api/v1/users`)

#### `GET /api/v1/users/me`
**Status:** ❌ Not implemented

**Description:** Get current user profile

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "user_id": "uuid-123",
  "email": "user@example.com",
  "username": "johndoe",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "status": "ACTIVE",
  "email_verified": true,
  "kyc_status": "APPROVED",
  "created_at": "2026-02-03T16:38:54+07:00",
  "last_login": "2026-02-03T16:38:54+07:00"
}
```

---

#### `PUT /api/v1/users/me`
**Status:** ❌ Not implemented

**Description:** Update user profile

**Request:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}
```

---

#### `POST /api/v1/users/me/verify-email`
**Status:** ❌ Not implemented

**Description:** Send email verification

---

#### `POST /api/v1/users/me/kyc`
**Status:** ❌ Not implemented

**Description:** Submit KYC documents

**Request:**
```json
{
  "document_type": "PASSPORT",
  "document_number": "AB123456",
  "document_image": "base64_encoded_image",
  "selfie_image": "base64_encoded_image"
}
```

---

### **4. TRADING ACCOUNTS** (`/api/v1/accounts`)

#### `GET /api/v1/accounts`
**Status:** ❌ Not implemented

**Description:** Get all user accounts

**Response:**
```json
{
  "accounts": [
    {
      "account_id": "uuid-456",
      "user_id": "uuid-123",
      "account_type": "CASH",
      "balance": 10000.00,
      "buying_power": 10000.00,
      "currency": "USD",
      "status": "ACTIVE",
      "created_at": "2026-02-03T16:38:54+07:00"
    },
    {
      "account_id": "uuid-789",
      "user_id": "uuid-123",
      "account_type": "MARGIN",
      "balance": 50000.00,
      "buying_power": 100000.00,
      "margin_used": 25000.00,
      "margin_available": 75000.00,
      "currency": "USD",
      "status": "ACTIVE",
      "created_at": "2026-02-03T16:38:54+07:00"
    }
  ]
}
```

---

#### `POST /api/v1/accounts`
**Status:** ❌ Not implemented

**Description:** Create new trading account

**Request:**
```json
{
  "account_type": "CASH",
  "currency": "USD"
}
```

---

#### `GET /api/v1/accounts/:account_id`
**Status:** ❌ Not implemented

**Description:** Get account details

---

#### `POST /api/v1/accounts/:account_id/deposit`
**Status:** ❌ Not implemented

**Description:** Deposit funds

**Request:**
```json
{
  "amount": 1000.00,
  "payment_method": "BANK_TRANSFER",
  "reference": "TXN123456"
}
```

---

#### `POST /api/v1/accounts/:account_id/withdraw`
**Status:** ❌ Not implemented

**Description:** Withdraw funds

**Request:**
```json
{
  "amount": 500.00,
  "bank_account": "1234567890",
  "bank_name": "ABC Bank"
}
```

---

### **5. ORDER MANAGEMENT** (`/api/v1/orders`)

#### `POST /api/v1/orders`
**Status:** ❌ Not implemented

**Description:** Submit new order

**Request:**
```json
{
  "account_id": "uuid-456",
  "symbol": "AAPL",
  "side": "BUY",
  "order_type": "LIMIT",
  "quantity": 10,
  "price": 150.00
}
```

**Response:**
```json
{
  "order_id": "uuid-order-1",
  "account_id": "uuid-456",
  "symbol": "AAPL",
  "side": "BUY",
  "order_type": "LIMIT",
  "quantity": 10,
  "price": 150.00,
  "status": "PENDING",
  "filled_quantity": 0,
  "avg_fill_price": 0,
  "created_at": "2026-02-03T16:38:54+07:00"
}
```

---

#### `GET /api/v1/orders`
**Status:** ❌ Not implemented

**Description:** Get all orders (with filters)

**Query Parameters:**
- `account_id` - Filter by account
- `symbol` - Filter by symbol
- `status` - Filter by status (PENDING, FILLED, CANCELLED)
- `limit` - Limit results (default: 50)
- `offset` - Pagination offset

**Response:**
```json
{
  "orders": [
    {
      "order_id": "uuid-order-1",
      "symbol": "AAPL",
      "side": "BUY",
      "order_type": "LIMIT",
      "quantity": 10,
      "price": 150.00,
      "status": "FILLED",
      "filled_quantity": 10,
      "avg_fill_price": 149.50,
      "created_at": "2026-02-03T16:38:54+07:00"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

---

#### `GET /api/v1/orders/:order_id`
**Status:** ❌ Not implemented

**Description:** Get order details

---

#### `DELETE /api/v1/orders/:order_id`
**Status:** ❌ Not implemented

**Description:** Cancel order

**Response:**
```json
{
  "order_id": "uuid-order-1",
  "status": "CANCELLED",
  "cancelled_at": "2026-02-03T16:38:54+07:00"
}
```

---

#### `PUT /api/v1/orders/:order_id`
**Status:** ❌ Not implemented

**Description:** Modify order (price/quantity)

**Request:**
```json
{
  "price": 151.00,
  "quantity": 15
}
```

---

### **6. PORTFOLIO & POSITIONS** (`/api/v1/portfolio`)

#### `GET /api/v1/portfolio`
**Status:** ❌ Not implemented

**Description:** Get portfolio summary

**Response:**
```json
{
  "account_id": "uuid-456",
  "total_value": 15000.00,
  "cash_balance": 5000.00,
  "positions_value": 10000.00,
  "total_pnl": 500.00,
  "total_pnl_percent": 3.45,
  "positions": [
    {
      "symbol": "AAPL",
      "quantity": 10,
      "avg_cost": 145.00,
      "current_price": 150.00,
      "total_cost": 1450.00,
      "current_value": 1500.00,
      "unrealized_pnl": 50.00,
      "unrealized_pnl_percent": 3.45
    }
  ]
}
```

---

#### `GET /api/v1/portfolio/positions`
**Status:** ❌ Not implemented

**Description:** Get all positions

---

#### `GET /api/v1/portfolio/positions/:symbol`
**Status:** ❌ Not implemented

**Description:** Get position for specific symbol

---

#### `GET /api/v1/portfolio/performance`
**Status:** ❌ Not implemented

**Description:** Get portfolio performance over time

**Query Parameters:**
- `period` - Time period (1D, 1W, 1M, 3M, 1Y, ALL)

**Response:**
```json
{
  "period": "1M",
  "start_value": 14500.00,
  "end_value": 15000.00,
  "total_return": 500.00,
  "total_return_percent": 3.45,
  "daily_returns": [
    {
      "date": "2026-02-01",
      "value": 14500.00,
      "return": 0
    },
    {
      "date": "2026-02-02",
      "value": 14600.00,
      "return": 100.00
    }
  ]
}
```

---

### **7. MARKET DATA** (`/api/v1/market`)

#### `GET /api/v1/market/stocks`
**Status:** ❌ Not implemented

**Description:** Get list of available stocks

**Query Parameters:**
- `search` - Search by symbol or name
- `exchange` - Filter by exchange
- `sector` - Filter by sector

**Response:**
```json
{
  "stocks": [
    {
      "symbol": "AAPL",
      "name": "Apple Inc.",
      "exchange": "NASDAQ",
      "sector": "Technology",
      "market_cap": 2800000000000,
      "is_tradable": true
    }
  ]
}
```

---

#### `GET /api/v1/market/stocks/:symbol`
**Status:** ❌ Not implemented

**Description:** Get stock details

---

#### `GET /api/v1/market/stocks/:symbol/price`
**Status:** ❌ Not implemented

**Description:** Get current price

**Response:**
```json
{
  "symbol": "AAPL",
  "price": 150.00,
  "bid": 149.95,
  "ask": 150.05,
  "volume": 50000000,
  "timestamp": "2026-02-03T16:38:54+07:00"
}
```

---

#### `GET /api/v1/market/stocks/:symbol/candles`
**Status:** ❌ Not implemented

**Description:** Get historical price data (OHLCV)

**Query Parameters:**
- `interval` - Candle interval (1m, 5m, 15m, 1h, 1d)
- `from` - Start timestamp
- `to` - End timestamp

**Response:**
```json
{
  "symbol": "AAPL",
  "interval": "1d",
  "candles": [
    {
      "timestamp": "2026-02-01T00:00:00Z",
      "open": 148.00,
      "high": 151.00,
      "low": 147.50,
      "close": 150.00,
      "volume": 50000000
    }
  ]
}
```

---

#### `GET /api/v1/market/stocks/:symbol/orderbook`
**Status:** ❌ Not implemented

**Description:** Get order book (market depth)

**Response:**
```json
{
  "symbol": "AAPL",
  "bids": [
    {
      "price": 149.95,
      "quantity": 1000,
      "orders": 5
    },
    {
      "price": 149.90,
      "quantity": 2000,
      "orders": 10
    }
  ],
  "asks": [
    {
      "price": 150.05,
      "quantity": 1500,
      "orders": 7
    },
    {
      "price": 150.10,
      "quantity": 2500,
      "orders": 12
    }
  ],
  "spread": 0.10,
  "mid_price": 150.00
}
```

---

### **8. TRADE HISTORY** (`/api/v1/trades`)

#### `GET /api/v1/trades`
**Status:** ❌ Not implemented

**Description:** Get trade history

**Query Parameters:**
- `account_id` - Filter by account
- `symbol` - Filter by symbol
- `from` - Start date
- `to` - End date
- `limit` - Limit results
- `offset` - Pagination offset

**Response:**
```json
{
  "trades": [
    {
      "trade_id": "uuid-trade-1",
      "order_id": "uuid-order-1",
      "symbol": "AAPL",
      "side": "BUY",
      "quantity": 10,
      "price": 149.50,
      "total_value": 1495.00,
      "timestamp": "2026-02-03T16:38:54+07:00"
    }
  ],
  "total": 1
}
```

---

#### `GET /api/v1/trades/:trade_id`
**Status:** ❌ Not implemented

**Description:** Get trade details

---

### **9. ADMIN OPERATIONS** (`/api/v1/admin`)

#### `GET /api/v1/admin/users`
**Status:** ❌ Not implemented

**Description:** List all users (admin only)

---

#### `PUT /api/v1/admin/users/:user_id/kyc`
**Status:** ❌ Not implemented

**Description:** Approve/reject KYC

**Request:**
```json
{
  "kyc_status": "APPROVED"
}
```

---

#### `GET /api/v1/admin/orders`
**Status:** ❌ Not implemented

**Description:** View all orders (admin only)

---

#### `GET /api/v1/admin/stats`
**Status:** ❌ Not implemented

**Description:** Get system statistics

**Response:**
```json
{
  "total_users": 1000,
  "active_users": 500,
  "total_orders": 10000,
  "total_trades": 8000,
  "total_volume": 5000000.00,
  "system_health": "healthy"
}
```

---

## 📊 ENDPOINT SUMMARY

| Category | Endpoints | Status |
|----------|-----------|--------|
| **Health** | 1 | ✅ 1 implemented |
| **Auth** | 4 | ❌ 0 implemented |
| **Users** | 4 | ❌ 0 implemented |
| **Accounts** | 5 | ❌ 0 implemented |
| **Orders** | 5 | ❌ 0 implemented |
| **Portfolio** | 4 | ❌ 0 implemented |
| **Market Data** | 5 | ❌ 0 implemented |
| **Trades** | 2 | ❌ 0 implemented |
| **Admin** | 4 | ❌ 0 implemented |
| **TOTAL** | **34** | **1/34 (3%)** |

---

## 🎯 IMPLEMENTATION PRIORITY

### **Phase 1: Core Authentication** (Week 1)
- [ ] POST /api/v1/auth/register
- [ ] POST /api/v1/auth/login
- [ ] POST /api/v1/auth/refresh
- [ ] GET /api/v1/users/me

### **Phase 2: Account Management** (Week 2)
- [ ] GET /api/v1/accounts
- [ ] POST /api/v1/accounts
- [ ] POST /api/v1/accounts/:id/deposit
- [ ] POST /api/v1/accounts/:id/withdraw

### **Phase 3: Order Management** (Week 3-4)
- [ ] POST /api/v1/orders
- [ ] GET /api/v1/orders
- [ ] GET /api/v1/orders/:id
- [ ] DELETE /api/v1/orders/:id

### **Phase 4: Portfolio & Market Data** (Week 5-6)
- [ ] GET /api/v1/portfolio
- [ ] GET /api/v1/market/stocks
- [ ] GET /api/v1/market/stocks/:symbol/price
- [ ] GET /api/v1/market/stocks/:symbol/orderbook

### **Phase 5: Advanced Features** (Week 7+)
- [ ] Trade history
- [ ] Performance analytics
- [ ] Admin operations
- [ ] Real-time WebSocket feeds

---

## 🔒 AUTHENTICATION

All endpoints (except `/health` and `/auth/*`) require authentication:

**Header:**
```
Authorization: Bearer <access_token>
```

**Error Response (401):**
```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired token"
}
```

---

## 📝 STANDARD ERROR RESPONSES

### **400 Bad Request**
```json
{
  "error": "Bad Request",
  "message": "Invalid input data",
  "details": {
    "field": "email",
    "issue": "Invalid email format"
  }
}
```

### **404 Not Found**
```json
{
  "error": "Not Found",
  "message": "Resource not found"
}
```

### **500 Internal Server Error**
```json
{
  "error": "Internal Server Error",
  "message": "An unexpected error occurred",
  "request_id": "abc123"
}
```

---

## 🚀 NEXT STEPS

To implement these endpoints, you need to:

1. **Create Repository Layer** - Database access
2. **Create Use Case Layer** - Business logic
3. **Create Handler Layer** - HTTP handlers
4. **Register Routes** - Wire everything together

**Would you like me to start implementing Phase 1 (Authentication)?** 🎯
