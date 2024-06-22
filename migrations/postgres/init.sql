-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(64) PRIMARY KEY
    -- Add other user-related columns as needed
);

-- Create characters table
CREATE TABLE IF NOT EXISTS characters (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    name VARCHAR(64) NOT NULL,
    -- Add other character-related columns as needed
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT characters_name_unique UNIQUE (name)
);

-- Create players table as an extension of characters
CREATE TABLE IF NOT EXISTS players (
    -- Using character_id as both primary key and foreign key since it's a 1:1 relationship
    character_id INT PRIMARY KEY,
    timestamp BIGINT NOT NULL,
    x FLOAT NOT NULL,
    y FLOAT NOT NULL,
    flipH BOOLEAN NOT NULL,
    hitpoints INT NOT NULL,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);
