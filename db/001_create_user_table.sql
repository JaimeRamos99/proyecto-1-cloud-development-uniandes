-- *******************************
-- * DROP EXISTING OBJECTS     *
-- *******************************

-- NOTE: Enable the following 4 lines if you need to completely drop tables 
--       and recreate them from scratch.

-- ***************************
-- * ENABLE EXTENSIONS       *
-- ***************************
-- CITEXT provides a case-insensitive text type used for the users.email column
-- so UNIQUE(email) treats 'User@x.com' and 'user@x.com' as the same value.
CREATE EXTENSION IF NOT EXISTS citext;

-- -----------------------------------------------------------------------------

-- ***************************
-- * CREATE USERS TABLE     *
-- ***************************
CREATE TABLE IF NOT EXISTS users (
    id             SERIAL       PRIMARY KEY, 
    first_name     TEXT         NOT NULL,
    last_name      TEXT         NOT NULL,
    email          CITEXT       NOT NULL UNIQUE,
    password_hash  TEXT         NOT NULL, 
    city           TEXT         NOT NULL,
    country        TEXT         NOT NULL 
);


COMMENT ON TABLE users IS 'Contains user information';

-- COLUMN COMMENTS
COMMENT ON COLUMN users.id            IS 'Unique user identifier';
COMMENT ON COLUMN users.first_name    IS 'User given name';
COMMENT ON COLUMN users.last_name     IS 'User family name';
COMMENT ON COLUMN users.email         IS 'User email (case-insensitive), unique login identifier';
COMMENT ON COLUMN users.password_hash IS 'User password hash (e.g., bcrypt/argon2), never store plaintext';
COMMENT ON COLUMN users.city          IS 'User city (free-form text)';
COMMENT ON COLUMN users.country       IS 'User country (free-form text)';

