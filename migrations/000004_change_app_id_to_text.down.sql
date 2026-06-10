ALTER TABLE refresh_tokens
ALTER COLUMN app_id TYPE UUID USING app_id::UUID;

ALTER TABLE refresh_tokens
ADD CONSTRAINT refresh_tokens_app_id_fkey
FOREIGN KEY (app_id)
REFERENCES registered_apps(id);