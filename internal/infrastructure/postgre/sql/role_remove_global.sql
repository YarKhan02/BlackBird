DELETE FROM user_global_roles
WHERE user_id = $1 AND role_id = (SELECT id FROM global_roles WHERE name = $2)
