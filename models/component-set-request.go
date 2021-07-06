package models

import (
	cmodels "github.com/jental/webbugs-common-go/models"
)

// ComponentSetRequest - struct for component set
type ComponentSetRequest struct {
	IsActive *bool
	Walls    []*cmodels.Cell
}
