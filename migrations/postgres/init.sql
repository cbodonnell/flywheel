-- Create players table
CREATE TABLE IF NOT EXISTS players (
    player_id bigint PRIMARY KEY,
    timestamp bigint NOT NULL,
    x float NOT NULL,
    y float NOT NULL
);
