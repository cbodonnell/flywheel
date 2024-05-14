-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY
    -- Add other user-related columns as needed
);

-- Create characters table
CREATE TABLE IF NOT EXISTS characters (
    id INTEGER PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    -- Add other character-related columns as needed
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT characters_name_unique UNIQUE (name)
);

-- Create players table as an extension of characters
CREATE TABLE IF NOT EXISTS players (
    -- Using character_id as both primary key and foreign key since it's a 1:1 relationship
    character_id INTEGER PRIMARY KEY,
    timestamp INTEGER NOT NULL,
    x REAL NOT NULL,
    y REAL NOT NULL,
    hitpoints INTEGER NOT NULL,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);
