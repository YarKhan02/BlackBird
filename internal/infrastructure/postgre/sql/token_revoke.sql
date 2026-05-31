UPDATE refresh_tokens SET revoked = TRUE, revoked_at = NOW() WHERE id = $1
