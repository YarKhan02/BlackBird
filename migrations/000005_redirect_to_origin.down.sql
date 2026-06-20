ALTER TABLE registered_apps
ADD COLUMN redirect_uris TEXT[];

-- restore as single-item array
UPDATE registered_apps
SET redirect_uris = ARRAY[origin];

ALTER TABLE registered_apps
ALTER COLUMN redirect_uris SET NOT NULL;

ALTER TABLE registered_apps
DROP CONSTRAINT registered_apps_origin_unique;

ALTER TABLE registered_apps
DROP COLUMN origin;