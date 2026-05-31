UPDATE users SET is_banned = TRUE, updated_at = NOW() WHERE id = $1
