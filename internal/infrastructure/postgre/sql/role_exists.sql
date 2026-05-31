SELECT EXISTS(SELECT 1 FROM global_roles WHERE name = $1)
