SELECT gr.name
FROM user_global_roles ugr
JOIN global_roles gr ON gr.id = ugr.role_id
WHERE ugr.user_id = $1
ORDER BY gr.id
