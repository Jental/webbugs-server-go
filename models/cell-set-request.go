package models

import (
	"fmt"
	"strconv"

	cmodels "github.com/jental/webbugs-common-go/models"

	"github.com/google/uuid"
)

// CellSetRequest - struct for on cell
type CellSetRequest struct {
	CellType  *cmodels.CellType
	PlayerID  uuid.UUID
	Component *cmodels.Component
	IsBase    *bool
}

func (request CellSetRequest) String() string {
	var cellTypeStr string
	if request.CellType == nil {
		cellTypeStr = "nil"
	} else {
		switch *request.CellType {
		case cmodels.CellTypeBug:
			cellTypeStr = "bug"
		case cmodels.CellTypeWall:
			cellTypeStr = "wall"
		default:
			cellTypeStr = "unknown"
		}
	}

	var componentStr string
	if request.Component == nil {
		componentStr = "nil"
	} else {
		componentStr = strconv.Itoa(int(request.Component.ID))
	}

	var isBaseStr string
	if request.IsBase == nil {
		isBaseStr = "nil"
	} else {
		if *request.IsBase {
			isBaseStr = "true"
		} else {
			isBaseStr = "false"
		}
	}

	return fmt.Sprintf("CellSetRequest:{ %v %v %v %v }", request.PlayerID, cellTypeStr, componentStr, isBaseStr)
}

// FillCellWithCellSetRequest - fills a cell with a request data
func FillCellWithCellSetRequest(cell *cmodels.Cell, request CellSetRequest) {
	if request.PlayerID != uuid.Nil {
		cell.PlayerID = request.PlayerID
	}
	if request.CellType != nil {
		cell.CellType = *request.CellType
	}
	if request.Component != nil {
		cell.Component = request.Component
	}
	if request.IsBase != nil {
		cell.IsBase = *request.IsBase
	}
}

// FromCellSetRequest - creates a new cell from CellSetRequest
func CellFromCellSetRequest(request CellSetRequest, crd cmodels.Coordinates) cmodels.Cell {
	newCell := cmodels.Cell{}
	newCell.Crd = crd

	FillCellWithCellSetRequest(&newCell, request)

	return newCell
}

// ApplyCellSetRequest - applies a CellSetRequest
func ApplyCellSetRequest(field *cmodels.Field, request *CellSetRequest, crd cmodels.Coordinates) error {
	if request != nil {
		cell, exists := field.GetWithExists(crd)
		if exists && cell != nil {
			if crd != cell.Crd {
				return fmt.Errorf("page: set: Unmatching coordinates: %v, %v", crd, cell.Crd)
			}

			FillCellWithCellSetRequest(cell, *request)
		} else {
			newCell := CellFromCellSetRequest(*request, crd)
			field.Set(crd, &newCell)
		}
	} else {
		field.Set(crd, nil)
	}

	return nil
}
