SELECT id, client_id, name, is_active, created_at
FROM registered_apps
WHERE client_id = $1