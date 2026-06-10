ALTER TABLE refresh_tokens
DROP CONSTRAINT refresh_tokens_app_id_fkey;

ALTER TABLE refresh_tokens
ALTER COLUMN app_id TYPE TEXT USING app_id::TEXT;