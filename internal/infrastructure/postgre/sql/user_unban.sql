UPDATE users SET is_banned = FALSE, updated_at = NOW() WHERE id = $1
