BEGIN;

DROP TABLE IF EXISTS post_categories;
DROP INDEX IF EXISTS idx_categories_slug;
DROP TABLE IF EXISTS categories;

COMMIT;
