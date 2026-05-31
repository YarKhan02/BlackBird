UPDATE users
SET email = $2, is_verified = $3, is_banned = $4, failed_attempts = $5, locked_until = $6, updated_at = NOW()
WHERE id = $1
