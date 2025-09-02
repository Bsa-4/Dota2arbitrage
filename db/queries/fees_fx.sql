-- name: UpsertFee :one
INSERT INTO fees(market_id, kind, percent_bps, fixed_minor)
VALUES ($1, $2, $3, $4)
ON CONFLICT (market_id, kind)
DO UPDATE SET percent_bps = EXCLUDED.percent_bps,
              fixed_minor = EXCLUDED.fixed_minor
RETURNING *;

-- name: UpsertFx :one
INSERT INTO fx_rates(base, quote, rate)
VALUES ($1, $2, $3)
ON CONFLICT (base, quote)
DO UPDATE SET rate = EXCLUDED.rate, as_of = now()
RETURNING *;
