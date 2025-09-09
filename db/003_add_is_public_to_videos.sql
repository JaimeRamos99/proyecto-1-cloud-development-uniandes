-- *******************************
-- * ADD is_public COLUMN TO VIDEOS *
-- *******************************

-- Add is_public column to videos table
ALTER TABLE videos ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT false;

-- Add comment for the new column
COMMENT ON COLUMN videos.is_public IS 'Whether the video is publicly accessible or private';

-- Add index for performance when filtering by is_public
CREATE INDEX IF NOT EXISTS idx_videos_is_public ON videos(is_public);

-- Add composite index for user_id and is_public for efficient querying
CREATE INDEX IF NOT EXISTS idx_videos_user_id_is_public ON videos(user_id, is_public);
