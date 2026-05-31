SELECT id, email, password_hash, is_verified, is_banned, failed_attempts, locked_until, created_at, updated_at
FROM users WHERE email = $1
