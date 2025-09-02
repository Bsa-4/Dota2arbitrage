-- name: ListMarkets :many
SELECT id, code, name, currency, created_at
FROM markets
ORDER BY id;

-- name: CreateMarket :one
INSERT INTO markets(code, name, currency)
VALUES ($1, $2, $3)
RETURNING id, code, name, currency, created_at;
