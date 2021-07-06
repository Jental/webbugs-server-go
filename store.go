package main

import (
	"webbugs-server/models"

	cmodels "github.com/jental/webbugs-common-go/models"

	"github.com/google/uuid"
)

// Store - stores all app data
type Store struct {
	field         *cmodels.Field
	components    *cmodels.Components
	players       map[uuid.UUID]*models.PlayerInfo
	subscribtions []func()
	eventQueue    chan *models.Event
	isLocked      bool
}

// NewStore - creates a new store
func NewStore(pageRadius uint) Store {
	field := cmodels.NewField(pageRadius)
	var components cmodels.Components
	return Store{
		field:         &field,
		components:    &components,
		players:       make(map[uuid.UUID]*models.PlayerInfo, 0),
		subscribtions: make([]func(), 0),
		eventQueue:    make(chan *models.Event),
		isLocked:      false,
	}
}
