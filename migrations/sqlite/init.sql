-- Create players table
CREATE TABLE IF NOT EXISTS players (
    user_id VARCHAR(64) PRIMARY KEY,
    timestamp INTEGER NOT NULL,
    x REAL NOT NULL,
    y REAL NOT NULL
);
