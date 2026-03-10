-- ============================================================================
-- Seed positions for user 9a8392cc-8138-499c-85f3-a300355b9cbc
-- Account: 6d5bfb4d-bad4-4c21-802e-3272c79a2db9 (CASH)
-- Run: psql -U <user> -d <db> -f scripts/seed_positions.sql
-- ============================================================================

BEGIN;

-- Ensure table exists (matches GORM PositionModel)
CREATE TABLE IF NOT EXISTS positions (
  id                     VARCHAR(36) PRIMARY KEY,
  user_id                VARCHAR(36) NOT NULL,
  account_id             VARCHAR(36) NOT NULL,
  symbol                 VARCHAR(10) NOT NULL,
  quantity               INT         NOT NULL,
  avg_price              DECIMAL(20,4) NOT NULL,
  current_price          DECIMAL(20,4),
  unrealized_pnl         DECIMAL(20,2),
  unrealized_pnl_percent DECIMAL(10,4),
  created_at             TIMESTAMPTZ NOT NULL,
  updated_at             TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_positions_user_id    ON positions(user_id);
CREATE INDEX IF NOT EXISTS idx_positions_account_id ON positions(account_id);
CREATE INDEX IF NOT EXISTS idx_positions_symbol     ON positions(symbol);

INSERT INTO positions (id, user_id, account_id, symbol, quantity, avg_price, current_price, unrealized_pnl, unrealized_pnl_percent, created_at, updated_at)
VALUES
  -- Technology
  ('b0000001-0000-0000-0000-000000000001', '9a8392cc-8138-499c-85f3-a300355b9cbc', '6d5bfb4d-bad4-4c21-802e-3272c79a2db9',
   'AAPL',  50,  178.5000, 185.2000,   335.00,  3.7535, NOW(), NOW()),

  ('b0000001-0000-0000-0000-000000000002', '9a8392cc-8138-499c-85f3-a300355b9cbc', '6d5bfb4d-bad4-4c21-802e-3272c79a2db9',
   'MSFT', 30,  380.0000, 395.5000,   465.00,  4.0789, NOW(), NOW()),

  ('b0000001-0000-0000-0000-000000000003', '9a8392cc-8138-499c-85f3-a300355b9cbc', '6d5bfb4d-bad4-4c21-802e-3272c79a2db9',
   'NVDA', 100,  85.2000,  92.4000,   720.00,  8.4507, NOW(), NOW()),

  ('b0000001-0000-0000-0000-000000000004', '9a8392cc-8138-499c-85f3-a300355b9cbc', '6d5bfb4d-bad4-4c21-802e-3272c79a2db9',
   'GOOGL', 20, 140.0000, 148.7500,   175.00,  6.2500, NOW(), NOW()),

  -- Finance
  ('b0000001-0000-0000-0000-000000000005', '9a8392cc-8138-499c-85f3-a300355b9cbc', '6d5bfb4d-bad4-4c21-802e-3272c79a2db9',
   'JPM',  40,  195.0000, 201.3000,   252.00,  3.2308, NOW(), NOW()),

  ('b0000001-0000-0000-0000-000000000006', '9a8392cc-8138-499c-85f3-a300355b9cbc', '6d5bfb4d-bad4-4c21-802e-3272c79a2db9',
   'V',    25,  275.0000, 282.5000,   187.50,  2.7273, NOW(), NOW()),

  -- Healthcare
  ('b0000001-0000-0000-0000-000000000007', '9a8392cc-8138-499c-85f3-a300355b9cbc', '6d5bfb4d-bad4-4c21-802e-3272c79a2db9',
   'JNJ',  60,  155.0000, 158.2000,   192.00,  2.0645, NOW(), NOW()),

  -- Energy
  ('b0000001-0000-0000-0000-000000000008', '9a8392cc-8138-499c-85f3-a300355b9cbc', '6d5bfb4d-bad4-4c21-802e-3272c79a2db9',
   'XOM',  80,  105.5000, 110.2000,   376.00,  4.4550, NOW(), NOW()),

  -- Consumer
  ('b0000001-0000-0000-0000-000000000009', '9a8392cc-8138-499c-85f3-a300355b9cbc', '6d5bfb4d-bad4-4c21-802e-3272c79a2db9',
   'KO',  120,   58.5000,  60.1000,   192.00,  2.7350, NOW(), NOW()),

  -- Entertainment
  ('b0000001-0000-0000-0000-000000000010', '9a8392cc-8138-499c-85f3-a300355b9cbc', '6d5bfb4d-bad4-4c21-802e-3272c79a2db9',
   'NFLX', 15,  620.0000, 645.0000,   375.00,  4.0323, NOW(), NOW())

ON CONFLICT (id) DO NOTHING;

COMMIT;
