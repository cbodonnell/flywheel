-- Create players table
CREATE TABLE IF NOT EXISTS players (
    player_id INTEGER PRIMARY KEY,
    timestamp INTEGER NOT NULL,
    x REAL NOT NULL,
    y REAL NOT NULL
);
