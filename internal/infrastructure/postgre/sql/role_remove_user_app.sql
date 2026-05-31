DELETE FROM user_app_roles
WHERE user_id = $1 AND app_id = $2 AND role = $3
