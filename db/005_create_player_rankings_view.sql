-- *******************************
-- * CREATE PLAYER RANKINGS VIEW *
-- *******************************

-- Create a materialized view for player rankings based on votes
-- This view will be refreshed every minute to provide dynamic rankings
CREATE MATERIALIZED VIEW IF NOT EXISTS player_rankings AS
SELECT 
    u.id AS user_id,
    u.first_name,
    u.last_name,
    u.email,
    u.city,
    u.country,
    COALESCE(vote_stats.total_votes, 0) AS total_votes,
    ROW_NUMBER() OVER (ORDER BY COALESCE(vote_stats.total_votes, 0) DESC, u.id ASC) AS ranking,
    NOW() AS last_updated
FROM users u
LEFT JOIN (
    SELECT 
        v.user_id,
        COUNT(vo.id) AS total_votes
    FROM videos v
    LEFT JOIN votes vo ON v.id = vo.video_id
    WHERE v.deleted_at IS NULL -- Only include non-deleted videos
    GROUP BY v.user_id
) vote_stats ON u.id = vote_stats.user_id
ORDER BY total_votes DESC, u.id ASC;

-- Create a unique index on the materialized view for better performance
CREATE UNIQUE INDEX IF NOT EXISTS idx_player_rankings_user_id ON player_rankings(user_id);
CREATE INDEX IF NOT EXISTS idx_player_rankings_total_votes ON player_rankings(total_votes DESC);
CREATE INDEX IF NOT EXISTS idx_player_rankings_ranking ON player_rankings(ranking);
CREATE INDEX IF NOT EXISTS idx_player_rankings_country ON player_rankings(country);
CREATE INDEX IF NOT EXISTS idx_player_rankings_city ON player_rankings(city);

COMMENT ON MATERIALIZED VIEW player_rankings IS 'Player rankings based on total votes received on their videos';

-- COLUMN COMMENTS
COMMENT ON COLUMN player_rankings.user_id IS 'Unique user identifier';
COMMENT ON COLUMN player_rankings.first_name IS 'User given name';
COMMENT ON COLUMN player_rankings.last_name IS 'User family name';
COMMENT ON COLUMN player_rankings.email IS 'User email';
COMMENT ON COLUMN player_rankings.city IS 'User city';
COMMENT ON COLUMN player_rankings.country IS 'User country';
COMMENT ON COLUMN player_rankings.total_votes IS 'Total number of votes received across all user videos';
COMMENT ON COLUMN player_rankings.ranking IS 'Current ranking position (1 is best)';
COMMENT ON COLUMN player_rankings.last_updated IS 'Timestamp when the view was last refreshed';

-- Create a function to refresh the materialized view
CREATE OR REPLACE FUNCTION refresh_player_rankings()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY player_rankings;
END;
$$ LANGUAGE plpgsql;

-- Schedule automatic refresh every minute using pg_cron extension
-- Note: pg_cron needs to be installed and configured in your PostgreSQL instance
-- If pg_cron is not available, you can use a cron job or scheduled task to call the refresh function
DO $$
BEGIN
    -- Check if pg_cron extension is available
    IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'pg_cron') THEN
        -- Create the extension if it doesn't exist
        CREATE EXTENSION IF NOT EXISTS pg_cron;
        
        -- Schedule the refresh job (every minute)
        -- Remove existing job if it exists to avoid duplicates
        PERFORM cron.unschedule('refresh-player-rankings');
        
        -- Schedule new job to refresh every minute
        PERFORM cron.schedule('refresh-player-rankings', '* * * * *', 'SELECT refresh_player_rankings();');
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        -- pg_cron might not be available, that's okay
        -- The view can still be refreshed manually or via external scheduler
        RAISE NOTICE 'pg_cron extension not available, manual refresh required for player_rankings view';
END
$$;

-- Initial refresh of the materialized view
SELECT refresh_player_rankings();
