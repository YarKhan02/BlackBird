INSERT INTO user_app_roles (user_id, app_id, role)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING
