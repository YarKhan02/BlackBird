SELECT 1
FROM registered_apps
WHERE orgin = $1
LIMIT 1;