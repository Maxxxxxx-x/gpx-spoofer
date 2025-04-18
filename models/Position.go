package models

type Position struct {
	Lat float64 `json:"latitude"`
	Lon float64 `json:"longitude"`
	Elv float64 `json:"elevation,omitempty"`
}
