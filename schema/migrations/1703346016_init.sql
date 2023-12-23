-- Create game_states table
CREATE TABLE IF NOT EXISTS game_states (
    timestamp bigint PRIMARY KEY
);

-- Create players table
CREATE TABLE IF NOT EXISTS players (
    player_id integer PRIMARY KEY
);

-- Create player_states table
CREATE TABLE IF NOT EXISTS player_states (
    timestamp bigint,
    player_id integer,
    -- Primary key constraint
    PRIMARY KEY (timestamp, player_id),
    -- Foreign key constraints
    FOREIGN KEY (timestamp) REFERENCES game_states(timestamp),
    FOREIGN KEY (player_id) REFERENCES players(player_id)
);

-- Create player_positions table
CREATE TABLE IF NOT EXISTS player_positions (
    timestamp bigint,
    player_id integer,
    x float NOT NULL,
    y float NOT NULL,
    -- Primary key constraint
    PRIMARY KEY (timestamp, player_id),
    -- Foreign key constraints
    FOREIGN KEY (timestamp, player_id) REFERENCES player_states(timestamp, player_id)
);
