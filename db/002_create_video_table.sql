-- *******************************
-- * CREATE VIDEOS TABLE        *
-- *******************************
CREATE TYPE video_status AS ENUM (
  'uploaded','processed'
);


CREATE TABLE IF NOT EXISTS videos (
    id             SERIAL           PRIMARY KEY, 
    title          TEXT             NOT NULL,
    status         video_status     NOT NULL DEFAULT 'uploaded',
    uploaded_at    TIMESTAMP        NOT NULL DEFAULT NOW(),
    processed_at   TIMESTAMP        NULL,
    deleted_at     TIMESTAMP        NULL,
    user_id        INTEGER          NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

COMMENT ON TABLE videos IS 'Contains video metadata and status information';

-- COLUMN COMMENTS
COMMENT ON COLUMN videos.id            IS 'Unique video identifier';
COMMENT ON COLUMN videos.title         IS 'Video title provided by user';
COMMENT ON COLUMN videos.status        IS 'Video status: uploaded, processing, processed, failed';
COMMENT ON COLUMN videos.uploaded_at   IS 'Timestamp when video was uploaded';
COMMENT ON COLUMN videos.processed_at  IS 'Timestamp when video processing completed (nullable)';
COMMENT ON COLUMN videos.deleted_at    IS 'Soft delete timestamp (nullable)';
COMMENT ON COLUMN videos.user_id       IS 'Foreign key reference to users table';

-- INDEXES
CREATE INDEX IF NOT EXISTS idx_videos_user_id ON videos(user_id);
