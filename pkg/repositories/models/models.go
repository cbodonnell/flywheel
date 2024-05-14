package models

type User struct {
	ID string `json:"id"`
}

type Character struct {
	ID     int32  `json:"id"`
	UserID string `json:"user_id,omitempty"`
	Name   string `json:"name"`
}

type Player struct {
	CharacterID int32   `json:"character_id"`
	Timestamp   int64   `json:"timestamp"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Hitpoints   int16   `json:"hitpoints"`
}
