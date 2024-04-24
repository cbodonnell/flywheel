-- Create players table
CREATE TABLE IF NOT EXISTS players (
    player_id VARCHAR(64) PRIMARY KEY,
    timestamp INTEGER NOT NULL,
    x REAL NOT NULL,
    y REAL NOT NULL
);
