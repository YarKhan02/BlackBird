SELECT id, client_id, client_secret, name, redirect_uris, is_active, created_at
FROM registered_apps
WHERE client_id = $1