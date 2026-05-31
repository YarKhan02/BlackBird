INSERT INTO user_global_roles (user_id, role_id)
SELECT $1, id FROM global_roles WHERE name = $2
ON CONFLICT DO NOTHING
