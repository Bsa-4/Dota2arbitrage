-- name: ListBots :many
SELECT id, name, vps_label, status, last_heartbeat_at, created_at
FROM bots
ORDER BY id;

-- name: CreateBot :one
INSERT INTO bots(name, vps_label, status, market_accounts_json)
VALUES ($1, $2, 'offline', COALESCE($3, '{}'::jsonb))
RETURNING id, name, vps_label, status, last_heartbeat_at, created_at;

-- name: HeartbeatBot :one
UPDATE bots
SET status = $2,
    last_heartbeat_at = now()
WHERE id = $1
RETURNING id, name, vps_label, status, last_heartbeat_at, created_at;
