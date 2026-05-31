INSERT INTO refresh_tokens (id, user_id, token_hash, app_id, user_agent, ip_address, expires_at, revoked)
VALUES ($1, $2, $3, $4, $5, $6, $7, FALSE)
RETURNING created_at
