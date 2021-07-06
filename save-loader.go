package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
	"webbugs-server/models"

	cmodels "github.com/jental/webbugs-common-go/models"

	"github.com/google/uuid"
)

// ClearField - clears data in store
func ClearField(store *Store) {
	prevIsLocked := store.isLocked
	store.isLocked = true
	defer func() { store.isLocked = prevIsLocked }()

	var components cmodels.Components
	store.components = &components
	store.players = make(map[uuid.UUID]*models.PlayerInfo)

	keys := make([]int64, 0)
	store.field.Grid.Range(func(key interface{}, value interface{}) bool {
		keys = append(keys, key.(int64))
		return true
	})
	for _, key := range keys {
		store.field.Grid.Store(key, nil)
	}
}

// LoadSave - loads a save file
func LoadSave(fileName string, store *Store) {
	prevIsLocked := store.isLocked
	store.isLocked = true
	defer func() { store.isLocked = prevIsLocked }()

	bytes, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return
	}

	ClearField(store)

	var data map[string]interface{}
	json.Unmarshal(bytes, &data)

	componentsDataI, exists := data["components"]
	if !exists {
		return
	}
	componentsData := componentsDataI.(map[string]interface{})
	for key, valueI := range componentsData {
		componentIDint, err := strconv.Atoi(key)
		if err == nil {
			componentID := uint(componentIDint)
			value := valueI.(map[string]interface{})

			isActiveI, exists := value["isActive"]
			isActive := false
			if exists {
				isActive = isActiveI.(bool)
			}

			newComponent := cmodels.Component{
				ID:       componentID,
				IsActive: isActive,
				Walls:    make([]*cmodels.Cell, 0),
			}
			store.components.Set(&newComponent)

			log.Printf("component: %v", newComponent)
		}
	}

	fieldDataI, exists := data["field"]
	if !exists {
		return
	}
	fieldData := fieldDataI.(map[string]interface{})

	log.Println("loaded data:")
	for key, valueI := range fieldData {
		if valueI != nil {
			value := valueI.(map[string]interface{})
			log.Printf("%v: %v", key, value)

			cellTypeI, exists := value["type"]
			if !exists {
				continue
			}
			cellTypeF := cellTypeI.(float64)
			var cellType cmodels.CellType
			switch cellTypeF {
			case 0:
				cellType = cmodels.CellTypeBug
			case 1:
				cellType = cmodels.CellTypeWall
			}

			playerIDStr, exists := value["playerID"]
			if !exists {
				continue
			}
			playerID, err := uuid.Parse(playerIDStr.(string))
			if err != nil {
				log.Print(err)
				continue
			}

			fcrdi, exists := value["p"]
			if !exists {
				continue
			}
			crdi, exists := fcrdi.(map[string]interface{})["cell"]
			if !exists {
				continue
			}
			crdm := crdi.(map[string]interface{})
			fcrd := cmodels.NewCoordinates(int64(crdm["x"].(float64)), int64(crdm["y"].(float64)), int64(crdm["z"].(float64)))

			isBaseI, exists := value["isBase"]
			var isBase bool
			if !exists {
				isBase = false
			} else {
				isBase = isBaseI.(bool)
			}

			componentIDI, exists := value["component_id"]
			var component *cmodels.Component = nil
			if exists && componentIDI != nil {
				componentIDint, err := strconv.Atoi(componentIDI.(string))
				componentID := uint(componentIDint)
				if err == nil {
					cmp, exists := store.components.Get(componentID)
					if exists {
						component = (*cmodels.Component)(cmp)
					} else {
						component = &cmodels.Component{
							ID:       componentID,
							Walls:    make([]*cmodels.Cell, 0),
							IsActive: false,
						}
						store.components.Set(component)
					}
				}
			}

			request := models.CellSetRequest{
				CellType:  &cellType,
				PlayerID:  playerID,
				Component: component,
				IsBase:    &isBase,
			}
			models.ApplyCellSetRequest(store.field, &request, fcrd)

			store.players[playerID] = &models.PlayerInfo{
				ID:           playerID,
				Name:         playerID.String(),
				Client:       nil,
				LastActivity: time.Now().UTC(),
			}

			if component != nil {
				wall := store.field.Get(fcrd)
				component.Walls = append(component.Walls, wall)
				if !component.IsActive {
					component.IsActive = store.field.CheckIfComponentActive(component, wall.PlayerID)
				}
			}

			log.Printf("%v", request)
		}
	}
}
