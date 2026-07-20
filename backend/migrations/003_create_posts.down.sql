BEGIN;

DROP INDEX IF EXISTS idx_posts_slug;
DROP INDEX IF EXISTS idx_posts_status;
DROP INDEX IF EXISTS idx_posts_author_id;
DROP TABLE IF EXISTS posts;
DROP TYPE IF EXISTS post_status;

COMMIT;
