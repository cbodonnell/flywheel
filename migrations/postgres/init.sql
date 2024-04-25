-- Create players table
CREATE TABLE IF NOT EXISTS players (
    user_id varchar(64) PRIMARY KEY,
    timestamp bigint NOT NULL,
    x float NOT NULL,
    y float NOT NULL
);
