BEGIN;

DROP TABLE IF EXISTS email_verification_tokens;
ALTER TABLE users DROP COLUMN IF EXISTS is_email_verified;

COMMIT;
