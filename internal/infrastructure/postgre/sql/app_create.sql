INSERT INTO registered_apps (id, client_id, client_secret, name, origin, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING created_at