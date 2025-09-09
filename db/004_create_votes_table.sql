-- *******************************
-- * CREATE VOTES TABLE           *
-- *******************************

CREATE TABLE IF NOT EXISTS votes (
    id         SERIAL    PRIMARY KEY,
    user_id    INTEGER   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id   INTEGER   NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    voted_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Ensure a user can only vote once per video
    UNIQUE(user_id, video_id)
);

COMMENT ON TABLE votes IS 'Contains user votes for videos';

-- COLUMN COMMENTS
COMMENT ON COLUMN votes.id        IS 'Unique vote identifier';
COMMENT ON COLUMN votes.user_id   IS 'Foreign key reference to users table';
COMMENT ON COLUMN votes.video_id  IS 'Foreign key reference to videos table';
COMMENT ON COLUMN votes.voted_at  IS 'Timestamp when vote was cast';

-- INDEXES for performance
CREATE INDEX IF NOT EXISTS idx_votes_user_id ON votes(user_id);
CREATE INDEX IF NOT EXISTS idx_votes_video_id ON votes(video_id);
CREATE INDEX IF NOT EXISTS idx_votes_user_video ON votes(user_id, video_id);
