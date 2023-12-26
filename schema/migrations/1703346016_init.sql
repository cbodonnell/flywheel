-- Create players table
CREATE TABLE IF NOT EXISTS players (
    player_id int PRIMARY KEY,
    created_at bigint NOT NULL,
    updated_at bigint,
    x float NOT NULL,
    y float NOT NULL
);
