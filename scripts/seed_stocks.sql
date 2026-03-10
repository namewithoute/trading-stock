-- ============================================================================
-- Seed data for stocks and prices tables
-- Run: psql -U <user> -d <db> -f scripts/seed_stocks.sql
-- ============================================================================

BEGIN;

-- ─── Stocks ─────────────────────────────────────────────────────────────────

INSERT INTO stocks (id, symbol, name, exchange, sector, industry, is_active, is_tradable, created_at, updated_at)
VALUES
  -- Technology
  ('a0000001-0000-0000-0000-000000000001', 'AAPL',  'Apple Inc.',                'NASDAQ', 'Technology',        'Consumer Electronics',     true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000002', 'MSFT',  'Microsoft Corporation',     'NASDAQ', 'Technology',        'Software - Infrastructure', true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000003', 'GOOGL', 'Alphabet Inc.',             'NASDAQ', 'Technology',        'Internet Content',         true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000004', 'AMZN',  'Amazon.com Inc.',           'NASDAQ', 'Technology',        'Internet Retail',          true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000005', 'META',  'Meta Platforms Inc.',       'NASDAQ', 'Technology',        'Internet Content',         true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000006', 'NVDA',  'NVIDIA Corporation',        'NASDAQ', 'Technology',        'Semiconductors',           true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000007', 'TSLA',  'Tesla Inc.',                'NASDAQ', 'Technology',        'Auto Manufacturers',       true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000008', 'TSM',   'Taiwan Semiconductor',      'NYSE',   'Technology',        'Semiconductors',           true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000009', 'AVGO',  'Broadcom Inc.',             'NASDAQ', 'Technology',        'Semiconductors',           true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000010', 'ORCL',  'Oracle Corporation',        'NYSE',   'Technology',        'Software - Infrastructure', true, true, NOW(), NOW()),

  -- Finance
  ('a0000001-0000-0000-0000-000000000011', 'JPM',   'JPMorgan Chase & Co.',      'NYSE',   'Financial Services','Banks - Diversified',      true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000012', 'V',     'Visa Inc.',                 'NYSE',   'Financial Services','Credit Services',          true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000013', 'MA',    'Mastercard Inc.',           'NYSE',   'Financial Services','Credit Services',          true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000014', 'BAC',   'Bank of America Corp.',     'NYSE',   'Financial Services','Banks - Diversified',      true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000015', 'GS',    'Goldman Sachs Group',       'NYSE',   'Financial Services','Capital Markets',          true, true, NOW(), NOW()),

  -- Healthcare
  ('a0000001-0000-0000-0000-000000000016', 'JNJ',   'Johnson & Johnson',         'NYSE',   'Healthcare',        'Drug Manufacturers',       true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000017', 'UNH',   'UnitedHealth Group',        'NYSE',   'Healthcare',        'Healthcare Plans',         true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000018', 'PFE',   'Pfizer Inc.',               'NYSE',   'Healthcare',        'Drug Manufacturers',       true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000019', 'ABBV',  'AbbVie Inc.',               'NYSE',   'Healthcare',        'Drug Manufacturers',       true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000020', 'LLY',   'Eli Lilly and Company',     'NYSE',   'Healthcare',        'Drug Manufacturers',       true, true, NOW(), NOW()),

  -- Energy
  ('a0000001-0000-0000-0000-000000000021', 'XOM',   'Exxon Mobil Corporation',   'NYSE',   'Energy',            'Oil & Gas Integrated',     true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000022', 'CVX',   'Chevron Corporation',       'NYSE',   'Energy',            'Oil & Gas Integrated',     true, true, NOW(), NOW()),

  -- Consumer
  ('a0000001-0000-0000-0000-000000000023', 'KO',    'The Coca-Cola Company',     'NYSE',   'Consumer Defensive','Beverages - Non-Alcoholic',true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000024', 'PEP',   'PepsiCo Inc.',              'NASDAQ', 'Consumer Defensive','Beverages - Non-Alcoholic',true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000025', 'WMT',   'Walmart Inc.',              'NYSE',   'Consumer Defensive','Discount Stores',          true, true, NOW(), NOW()),

  -- Industrial / Other
  ('a0000001-0000-0000-0000-000000000026', 'DIS',   'The Walt Disney Company',   'NYSE',   'Communication',     'Entertainment',            true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000027', 'NFLX',  'Netflix Inc.',              'NASDAQ', 'Communication',     'Entertainment',            true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000028', 'BA',    'The Boeing Company',        'NYSE',   'Industrials',       'Aerospace & Defense',      true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000029', 'INTC',  'Intel Corporation',         'NASDAQ', 'Technology',        'Semiconductors',           true, true, NOW(), NOW()),
  ('a0000001-0000-0000-0000-000000000030', 'AMD',   'Advanced Micro Devices',    'NASDAQ', 'Technology',        'Semiconductors',           true, true, NOW(), NOW())
ON CONFLICT (symbol) DO NOTHING;

-- ─── Prices (latest snapshot per symbol) ────────────────────────────────────

INSERT INTO prices (id, symbol, price, bid, ask, volume, timestamp)
VALUES
  -- Technology
  (gen_random_uuid(), 'AAPL',  242.5000, 242.4500, 242.5500,  52340000, NOW()),
  (gen_random_uuid(), 'MSFT',  415.3000, 415.2500, 415.3500,  28100000, NOW()),
  (gen_random_uuid(), 'GOOGL', 178.9200, 178.8800, 178.9600,  31200000, NOW()),
  (gen_random_uuid(), 'AMZN',  205.7500, 205.7000, 205.8000,  45600000, NOW()),
  (gen_random_uuid(), 'META',  595.2000, 595.1500, 595.2500,  19800000, NOW()),
  (gen_random_uuid(), 'NVDA',  890.4000, 890.3000, 890.5000,  61000000, NOW()),
  (gen_random_uuid(), 'TSLA',  248.6000, 248.5000, 248.7000,  89200000, NOW()),
  (gen_random_uuid(), 'TSM',   168.3000, 168.2500, 168.3500,  17500000, NOW()),
  (gen_random_uuid(), 'AVGO',  192.4500, 192.4000, 192.5000,  10200000, NOW()),
  (gen_random_uuid(), 'ORCL',  175.8000, 175.7500, 175.8500,  12300000, NOW()),

  -- Finance
  (gen_random_uuid(), 'JPM',   245.2000, 245.1500, 245.2500,  14700000, NOW()),
  (gen_random_uuid(), 'V',     310.5000, 310.4500, 310.5500,   8200000, NOW()),
  (gen_random_uuid(), 'MA',    485.7500, 485.7000, 485.8000,   5600000, NOW()),
  (gen_random_uuid(), 'BAC',    42.3500,  42.3300,  42.3700,  38500000, NOW()),
  (gen_random_uuid(), 'GS',    568.9000, 568.8500, 568.9500,   3200000, NOW()),

  -- Healthcare
  (gen_random_uuid(), 'JNJ',   158.2000, 158.1500, 158.2500,  11000000, NOW()),
  (gen_random_uuid(), 'UNH',   527.4000, 527.3500, 527.4500,   4800000, NOW()),
  (gen_random_uuid(), 'PFE',    28.7500,  28.7300,  28.7700,  42000000, NOW()),
  (gen_random_uuid(), 'ABBV',  182.6000, 182.5500, 182.6500,   7100000, NOW()),
  (gen_random_uuid(), 'LLY',   785.3000, 785.2000, 785.4000,   5300000, NOW()),

  -- Energy
  (gen_random_uuid(), 'XOM',   115.8000, 115.7500, 115.8500,  16200000, NOW()),
  (gen_random_uuid(), 'CVX',   160.4500, 160.4000, 160.5000,  10800000, NOW()),

  -- Consumer
  (gen_random_uuid(), 'KO',     62.3500,  62.3300,  62.3700,  15900000, NOW()),
  (gen_random_uuid(), 'PEP',   172.8000, 172.7500, 172.8500,   6700000, NOW()),
  (gen_random_uuid(), 'WMT',    85.2000,  85.1800,  85.2200,  12400000, NOW()),

  -- Other
  (gen_random_uuid(), 'DIS',   112.5000, 112.4500, 112.5500,  14300000, NOW()),
  (gen_random_uuid(), 'NFLX',  720.6000, 720.5000, 720.7000,   8900000, NOW()),
  (gen_random_uuid(), 'BA',    182.3000, 182.2500, 182.3500,   9600000, NOW()),
  (gen_random_uuid(), 'INTC',   24.5500,  24.5300,  24.5700,  56700000, NOW()),
  (gen_random_uuid(), 'AMD',   178.9000, 178.8500, 178.9500,  43200000, NOW());

COMMIT;
