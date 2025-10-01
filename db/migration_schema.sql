-- MySQL Schema for Video Application with Job Queue

-- Drop existing tables if recreating (uncomment if needed)
-- DROP TABLE IF EXISTS votes;
-- DROP TABLE IF EXISTS video_jobs;
-- DROP TABLE IF EXISTS videos;
-- DROP TABLE IF EXISTS users;

-- *******************************
-- * CREATE USERS TABLE          *
-- *******************************
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    country VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- *******************************
-- * CREATE VIDEOS TABLE         *
-- *******************************
CREATE TABLE IF NOT EXISTS videos (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500),
    processed_path VARCHAR(500),
    status ENUM('uploaded', 'processing', 'processed', 'failed') NOT NULL DEFAULT 'uploaded',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_videos_user_id (user_id),
    INDEX idx_videos_status (status),
    INDEX idx_videos_is_public (is_public),
    INDEX idx_videos_user_id_is_public (user_id, is_public),
    INDEX idx_videos_file_path (file_path(255)),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- *******************************
-- * CREATE VIDEO_JOBS TABLE     *
-- *******************************
CREATE TABLE IF NOT EXISTS video_jobs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    video_id BIGINT NOT NULL,
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
    file_path VARCHAR(500) NOT NULL,
    output_path VARCHAR(500) NULL,
    error_message TEXT NULL,
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    
    INDEX idx_status (status),
    INDEX idx_video_id (video_id),
    INDEX idx_created_at (created_at),
    INDEX idx_status_attempts (status, attempts),
    
    FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- *******************************
-- * CREATE VOTES TABLE          *
-- *******************************
CREATE TABLE IF NOT EXISTS votes (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    video_id BIGINT NOT NULL,
    voted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_user_video (user_id, video_id),
    INDEX idx_votes_user_id (user_id),
    INDEX idx_votes_video_id (video_id),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- *******************************
-- * CREATE PLAYER_RANKINGS VIEW *
-- *******************************
CREATE OR REPLACE VIEW player_rankings AS
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
    WHERE v.deleted_at IS NULL
    GROUP BY v.user_id
) vote_stats ON u.id = vote_stats.user_id
ORDER BY total_votes DESC, u.id ASC;