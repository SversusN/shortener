BEGIN TRANSACTION
CREATE TABLE IF NOT EXISTS URLS
(short_url varchar(100) NOT NULL,
 original_url varchar(1000) NOT NULL,
 user_id uuid,
 is_deleted BOOL default FALSE);
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_original ON URLS(original_url) WHERE is_deleted = FALSE;
COMMIT TRANSACTION