package models

import (
	"fmt"

	cmodels "github.com/jental/webbugs-common-go/models"

	"github.com/google/uuid"
)

// EventType - type of an event
type EventType uint

// Available event types
const (
	EventTypeClick                   EventType = 1
	EventTypeSetBug                  EventType = 2
	EventTypeSetWall                 EventType = 3
	EventTypeUpdateComponentActivity EventType = 4
	EventTypeClearCell               EventType = 5
)

// Event - interface for events
type Event interface {
	String() string
}

// ClickEvent - click event
type ClickEvent struct {
	EventType EventType
	Crd       cmodels.Coordinates
	PlayerID  uuid.UUID
}

// NewClickEvent - creates new click event
func NewClickEvent(crd cmodels.Coordinates, playerID uuid.UUID) ClickEvent {
	return ClickEvent{
		EventType: EventTypeClick,
		Crd:       crd,
		PlayerID:  playerID,
	}
}

func (event ClickEvent) String() string {
	return fmt.Sprintf("ClickEvent:{ %v %v }", event.Crd, event.PlayerID)
}

// SetBugEvent - set bug event
type SetBugEvent struct {
	EventType EventType
	Crd       cmodels.Coordinates
	PlayerID  uuid.UUID
	IsBase    bool
}

// NewSetBugEvent - creates new set bug event
func NewSetBugEvent(crd cmodels.Coordinates, playerID uuid.UUID, isBase bool) SetBugEvent {
	return SetBugEvent{
		EventType: EventTypeSetBug,
		Crd:       crd,
		PlayerID:  playerID,
		IsBase:    isBase,
	}
}

func (event SetBugEvent) String() string {
	return fmt.Sprintf("SetBugEvent:{ %v %v %v }", event.Crd, event.PlayerID, event.IsBase)
}

// SetWallEvent - set wall event
type SetWallEvent struct {
	EventType EventType
	Crd       cmodels.Coordinates
	PlayerID  uuid.UUID
}

// NewSetWallEvent - creates new set wall event
func NewSetWallEvent(crd cmodels.Coordinates, playerID uuid.UUID) SetWallEvent {
	return SetWallEvent{
		EventType: EventTypeSetWall,
		Crd:       crd,
		PlayerID:  playerID,
	}
}

func (event SetWallEvent) String() string {
	return fmt.Sprintf("SetWallEvent:{ %v %v }", event.Crd, event.PlayerID)
}

// UpdateComponentActivityEvent - update component activity event. component activity should be updated based on current field state
type UpdateComponentActivityEvent struct {
	EventType EventType
	Component *cmodels.Component
}

// NewUpdateComponentActivityEvent - creates new click event
func NewUpdateComponentActivityEvent(component *cmodels.Component) UpdateComponentActivityEvent {
	return UpdateComponentActivityEvent{
		EventType: EventTypeUpdateComponentActivity,
		Component: component,
	}
}

func (event UpdateComponentActivityEvent) String() string {
	return fmt.Sprintf("UpdateComponentActivityEvent:{ %v }", event.Component.ID)
}

// ClearCellsEvent - cell clear event
type ClearCellsEvent struct {
	EventType EventType
	Crd       []cmodels.Coordinates
}

// NewClearCellsEvent - creates new clear cell event
func NewClearCellsEvent(crd []cmodels.Coordinates) ClearCellsEvent {
	return ClearCellsEvent{
		EventType: EventTypeClearCell,
		Crd:       crd,
	}
}

func (event ClearCellsEvent) String() string {
	return fmt.Sprintf("ClearCellEvent:{ %v }", event.Crd)
}
