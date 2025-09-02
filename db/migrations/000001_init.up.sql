-- доменные таблицы и индексы (UP)

CREATE TABLE IF NOT EXISTS markets (
  id            SERIAL PRIMARY KEY,
  code          TEXT NOT NULL UNIQUE,
  name          TEXT NOT NULL,
  currency      TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS fees (
  id            SERIAL PRIMARY KEY,
  market_id     INT NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
  kind          TEXT NOT NULL CHECK (kind IN ('buy','sell')),
  percent_bps   INT  NOT NULL DEFAULT 0,
  fixed_minor   BIGINT NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (market_id, kind)
);

CREATE TABLE IF NOT EXISTS fx_rates (
  id            SERIAL PRIMARY KEY,
  base          TEXT NOT NULL,
  quote         TEXT NOT NULL,
  rate          NUMERIC(20,8) NOT NULL,
  as_of         TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (base, quote)
);

CREATE TABLE IF NOT EXISTS bots (
  id            SERIAL PRIMARY KEY,
  name          TEXT NOT NULL,
  vps_label     TEXT NOT NULL,
  status        TEXT NOT NULL DEFAULT 'offline' CHECK (status IN ('offline','online','error')),
  last_heartbeat_at TIMESTAMPTZ,
  market_accounts_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS bots_status_idx ON bots(status);

CREATE TABLE IF NOT EXISTS items (
  id                 BIGSERIAL PRIMARY KEY,
  bot_id             INT REFERENCES bots(id) ON DELETE SET NULL,
  app_id             INT NOT NULL,
  asset_id           TEXT NOT NULL,
  market_hash_name   TEXT NOT NULL,
  state              TEXT NOT NULL DEFAULT 'free' CHECK (state IN ('free','listed','reserved','in_transfer')),
  location_market_id INT REFERENCES markets(id) ON DELETE SET NULL,
  tradable_at        TIMESTAMPTZ,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS items_state_idx ON items(state);
CREATE INDEX IF NOT EXISTS items_name_idx  ON items(market_hash_name);

CREATE TABLE IF NOT EXISTS orders (
  id              BIGSERIAL PRIMARY KEY,
  market_id       INT NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
  bot_id          INT REFERENCES bots(id) ON DELETE SET NULL,
  side            TEXT NOT NULL CHECK (side IN ('buy','sell')),
  price_minor     BIGINT NOT NULL,
  currency        TEXT NOT NULL,
  qty             INT NOT NULL DEFAULT 1,
  state           TEXT NOT NULL DEFAULT 'created' CHECK (state IN ('created','placed','filled','cancelled','failed')),
  idem_key        TEXT NOT NULL,
  ext_id          TEXT,
  item_key        TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS orders_idem_key_idx ON orders(idem_key);
CREATE INDEX IF NOT EXISTS orders_state_idx     ON orders(state);
CREATE INDEX IF NOT EXISTS orders_item_key_idx  ON orders(item_key);
CREATE INDEX IF NOT EXISTS orders_market_idx    ON orders(market_id);

CREATE TABLE IF NOT EXISTS trades (
  id               BIGSERIAL PRIMARY KEY,
  buy_order_id     BIGINT REFERENCES orders(id) ON DELETE SET NULL,
  sell_order_id    BIGINT REFERENCES orders(id) ON DELETE SET NULL,
  state            TEXT NOT NULL DEFAULT 'init' CHECK (state IN ('init','in_progress','settled','failed')),
  pnl_net_minor    BIGINT,
  hold_time_sec    INT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS opportunities (
  id              BIGSERIAL PRIMARY KEY,
  item_key        TEXT NOT NULL,
  buy_market_id   INT NOT NULL REFERENCES markets(id),
  sell_market_id  INT NOT NULL REFERENCES markets(id),
  spread_bps      INT NOT NULL,
  calc_json       JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS opportunities_item_idx   ON opportunities(item_key);
CREATE INDEX IF NOT EXISTS opportunities_spread_idx ON opportunities(spread_bps);

CREATE TABLE IF NOT EXISTS risk_limits (
  id              SERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  config_json     JSONB NOT NULL,
  enabled         BOOLEAN NOT NULL DEFAULT true,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
  id              SERIAL PRIMARY KEY,
  username        TEXT UNIQUE,
  trade_url_enc   TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
