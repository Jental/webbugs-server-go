package models

import (
	"time"

	gosocketio "github.com/ambelovsky/gosf-socketio"
	"github.com/google/uuid"
)

// PlayerInfo - information about player
type PlayerInfo struct {
	ID           uuid.UUID
	Name         string
	Client       *gosocketio.Channel
	LastActivity time.Time
}
