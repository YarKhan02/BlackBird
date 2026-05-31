UPDATE users SET failed_attempts = $2, locked_until = $3, updated_at = NOW() WHERE id = $1
