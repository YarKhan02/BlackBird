SELECT id, user_id, token_hash, app_id, user_agent, ip_address, expires_at, revoked, revoked_at, created_at
FROM refresh_tokens WHERE user_id = $1 ORDER BY created_at DESC
