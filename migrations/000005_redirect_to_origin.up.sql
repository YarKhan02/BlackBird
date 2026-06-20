ALTER TABLE registered_apps
ADD COLUMN origin TEXT;

-- migrate existing data (take first element from array)
UPDATE registered_apps
SET origin = redirect_uris[1];

ALTER TABLE registered_apps
ALTER COLUMN origin SET NOT NULL;

ALTER TABLE registered_apps
ADD CONSTRAINT registered_apps_origin_unique UNIQUE (origin);

ALTER TABLE registered_apps
DROP COLUMN redirect_uris;