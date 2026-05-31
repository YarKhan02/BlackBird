SELECT a.client_id, uar.role
FROM user_app_roles uar
JOIN registered_apps a ON a.id = uar.app_id
WHERE uar.user_id = $1
ORDER BY a.client_id
