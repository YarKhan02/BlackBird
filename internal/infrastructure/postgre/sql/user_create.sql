INSERT INTO users (id, email, password_hash, is_verified, is_banned, failed_attempts, locked_until)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING created_at, updated_at
