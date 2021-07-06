package main

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"
	"webbugs-server/models"

	cmodels "github.com/jental/webbugs-common-go/models"

	"github.com/google/uuid"
)

const SMALL_TIMEOUT time.Duration = 50 * time.Millisecond
const BIG_TIMEOUT time.Duration = 5 * time.Second

var logger *log.Logger

func getAngle(cell0 cmodels.Coordinates, cell1 cmodels.Coordinates) float64 {
	dx := cell1.X - cell0.X
	dy := cell1.Y - cell0.Y
	dz := cell1.Z - cell0.Z

	dxy := dx - dy
	if dxy == 0 {
		if dz > 0 {
			return 270
		} else {
			return 90
		}
	}

	tg := -float64(dz) / float64(dxy)

	if dxy >= 0 && tg >= 0 {
		return tg * 60 // I
	} else if dxy < 0 && tg < 0 {
		return 180 + tg*60 // II
	} else if dxy < 0 && tg >= 0 {
		return 180 + tg*60 // III
	} else {
		return 360 + tg*60 // IV
	}
}

var sq3 float64 = math.Sqrt(3)

func getCellCoordinatesByAngle(cell cmodels.Coordinates, angle float64) cmodels.Coordinates {
	rangle := math.Pi / 180 * angle // in radians

	dx := int64(math.Round(math.Sin(rangle+math.Pi/3) * 2 / sq3))
	dy := int64(math.Round(math.Sin(rangle-math.Pi/3) * 2 / sq3))
	dz := -int64(math.Round(math.Sin(rangle) * 2 / sq3))

	// logger.Println(cell, angle, dx, dy, dz)

	return cmodels.NewCoordinates(cell.X+dx, cell.Y+dy, cell.Z+dz)
}

func getOrt(cell0 cmodels.Coordinates, cell1 cmodels.Coordinates, direction bool) cmodels.Coordinates {
	angle := getAngle(cell0, cell1)
	// logger.Println(cell0, cell1, angle)

	var ortAngle float64
	if direction {
		// right direction
		ortAngle = math.Remainder(angle-60, 360)
	} else {
		// left direction
		ortAngle = math.Remainder(angle+60, 360)
	}

	return getCellCoordinatesByAngle(cell0, ortAngle)
}

// Identifies loop direction.
// false - to left  while moving from the start point
// true  - to right while moving from the start point
func getLoopDirection(loop []*cmodels.Cell) (bool, error) {
	if len(loop) < 4 {
		return false, errors.New("Loop is too small")
	}

	if loop[0] != loop[len(loop)-1] {
		return false, errors.New("Loop is invalid")
	}

	startAngle := getAngle(loop[0].Crd, loop[1].Crd)
	var totalDiff float64 = 0
	var totalAngleDiff float64 = 0
	var countLeft = 0
	var countRight = 0
	var lastTotalDiffBeforeZero float64 = 0
	for i, cell := range loop {
		if i == 0 || i >= len(loop)-2 {
			continue
		}

		absoluteAngle := getAngle(cell.Crd, loop[i+1].Crd)
		previousAngle := getAngle(loop[i-1].Crd, cell.Crd)
		relativeToStartAngle := math.Mod(360+absoluteAngle-startAngle, 360)
		relativeToPreviousAngle := math.Mod(360+absoluteAngle-previousAngle, 360)
		if relativeToPreviousAngle >= 180 {
			relativeToPreviousAngle = relativeToPreviousAngle - 360
		}
		rangle := math.Pi / 180 * relativeToStartAngle // in radians
		logger.Println("direction angle:", cell.Crd, loop[i+1].Crd, absoluteAngle, relativeToStartAngle, relativeToPreviousAngle)

		diff := math.Sin(rangle) / sq3 * 2
		totalDiff = totalDiff + diff
		totalAngleDiff = totalAngleDiff + relativeToPreviousAngle

		if math.Round(totalDiff) == 0 && math.Round(diff) != 0 {
			lastTotalDiffBeforeZero = totalDiff - diff
		}

		if diff < 0 {
			countRight = countRight + 1
		} else if diff > 0 {
			countLeft = countLeft + 1
		}

		logger.Println("direction diff:", diff, totalDiff, lastTotalDiffBeforeZero, totalAngleDiff)
	}

	var result0 bool
	if math.Round(totalDiff) == 0 {
		result0 = lastTotalDiffBeforeZero < 0
	} else {
		result0 = totalDiff < 0
	}

	var result1 = countLeft < countRight

	var result2 = totalAngleDiff < 0

	logger.Println("direction results:", result0, result1, result2)

	return result2, nil

	// if math.Abs(float64(countLeft)-float64(countRight)) < 5 {
	// 	return result0, errors.New("Uncertain direction result")
	// }

	// if result0 == result1 {
	// 	return result0, nil
	// } else {
	// 	return result0, errors.New("Unmatching direction results")
	// }
}

func findCellsInsideLoop(loop []*cmodels.Cell) ([]cmodels.Coordinates, error) {
	if loop[0] != loop[len(loop)-1] {
		return nil, errors.New("Loop is invalid")
	}

	direction, err := getLoopDirection(loop)
	if err != nil {
		return nil, err
	}

	insideCells := make([]cmodels.Coordinates, 0)
	newCells := make([]cmodels.Coordinates, 0)
	prevCells := make([]cmodels.Coordinates, len(loop))
	for i, cell := range loop {
		prevCells[i] = cell.Crd
	}

	for {
		logger.Println("next inside level", prevCells)
		for i, cell := range prevCells {
			if i < len(prevCells)-1 {
				ncellCrd := getOrt(cell, prevCells[i+1], direction)
				found := false
				for _, nc := range newCells {
					if nc.X == ncellCrd.X && nc.Y == ncellCrd.Y && nc.Z == ncellCrd.Z {
						found = true
						break
					}
				}
				if !found {
					for _, nc := range insideCells {
						if nc.X == ncellCrd.X && nc.Y == ncellCrd.Y && nc.Z == ncellCrd.Z {
							found = true
							break
						}
					}
				}
				if !found {
					for _, nc := range loop {
						if nc.Crd.X == ncellCrd.X && nc.Crd.Y == ncellCrd.Y && nc.Crd.Z == ncellCrd.Z {
							found = true
							break
						}
					}
				}
				if !found {
					newCells = append(newCells, ncellCrd)
				}
			}
		}

		if len(newCells) == 0 {
			break
		} else {
			insideCells = append(insideCells, newCells...)
			prevCells = newCells
			newCells = make([]cmodels.Coordinates, 0)
		}
	}

	return insideCells, nil
}

func (store *Store) populateLoop(loop []*cmodels.Cell, playerID uuid.UUID) ([]cmodels.Coordinates, error) {
	if loop[0] != loop[len(loop)-1] {
		return nil, errors.New("Loop is invalid")
	}

	allCells := make([]cmodels.Coordinates, len(loop))
	for i, cell := range loop {
		allCells[i] = cell.Crd
	}

	for {
		newCells := make([]cmodels.Coordinates, 0)
		for _, cell0 := range allCells {
			for _, cell1 := range allCells {
				if cell0 != cell1 && cmodels.AreNeighbours(cell0, cell1) {
					ort0 := getOrt(cell0, cell1, true)
					ortCell0 := store.field.Get(ort0)
					if ortCell0 != nil && ortCell0.CellType == cmodels.CellTypeWall && ortCell0.PlayerID == playerID {
						if !existsCrd(newCells, ort0) && !existsCrd(allCells, ort0) {
							newCells = append(newCells, ort0)
						}
					}
					ort1 := getOrt(cell0, cell1, false)
					ortCell1 := store.field.Get(ort1)
					if ortCell1 != nil && ortCell1.CellType == cmodels.CellTypeWall && ortCell1.PlayerID == playerID {
						if !existsCrd(newCells, ort1) && !existsCrd(allCells, ort1) {
							newCells = append(newCells, ort1)
						}
					}
				}
			}
		}

		if len(newCells) == 0 {
			break
		}

		logger.Println("populate: new cells:", newCells)
		allCells = append(allCells, newCells...)
	}

	return allCells, nil
}

func (store *Store) findNextLoopStart(component *cmodels.Component) *cmodels.Cell {
	wallsWithOtherNeighbours := make([]*cmodels.Cell, 0)
	for _, wall := range component.Walls {
		neighbours := store.field.GetNeibhours(wall.Crd)

		otherNeighboursPresent := false
		for _, n := range neighbours {
			// if n == nil {
			// 	logger.Println("findNextLoopStart: nil neighbour")
			// } else {
			// 	logger.Printf("findNextLoopStart: neighbour [%v]: %v %v", wall.Crd, n.CellType, n.PlayerID)
			// }
			if n == nil ||
				n.CellType == cmodels.CellTypeBug ||
				(n.CellType == cmodels.CellTypeWall && n.PlayerID != wall.PlayerID) {
				otherNeighboursPresent = true
				break
			}
		}

		if otherNeighboursPresent {
			wallsWithOtherNeighbours = append(wallsWithOtherNeighbours, wall)
		}
	}

	if len(wallsWithOtherNeighbours) == 0 {
		return nil
	}

	wallIdx := int(rand.Int31n(int32(len(wallsWithOtherNeighbours))))
	wall := wallsWithOtherNeighbours[wallIdx]

	return wall
}

func processComponent(component *cmodels.Component) bool {
	start := store.findNextLoopStart(component)
	// var start *cmodels.Cell = store.field.Get(cmodels.NewCoordinates(0, 2, -2)) // circle_wall_1.json
	//var start *cmodels.Cell = store.field.Get(cmodels.NewCoordinates(8, -9, 1)) // circle_wall_0.json
	//var start *cmodels.Cell = store.field.Get(cmodels.NewCoordinates(-1, -4, 5)) // circle_wall_right_0.json
	// var start *cmodels.Cell = store.field.Get(cmodels.NewCoordinates(5, -7, 2)) // circle_wall_right_1.json
	// var start *cmodels.Cell = store.field.Get(cmodels.NewCoordinates(5, 0, -5)) // circle_wall_1.json
	// var start *cmodels.Cell = store.field.Get(cmodels.NewCoordinates(3, -2, -1)) // circle_wall_1.json
	// var start *cmodels.Cell = store.field.Get(cmodels.NewCoordinates(4, -1, -3)) // circle_wall_2.json
	if start == nil {
		logger.Println("loop start not found")
		return false
	}
	logger.Println("loop start:", *start)

	loop, found := store.findWallLoop(start.PlayerID, start)
	if !found {
		logger.Println("loop not found")
		return false
	}

	logger.Println("loop:")
	for _, c := range loop {
		logger.Println(c.Crd)
	}

	direction, err := getLoopDirection(loop)
	if err != nil {
		logger.Println("failed to get loop direction", err)
		return false
	}

	if direction {
		logger.Println("right loop")
	} else {
		logger.Println("left loop")
	}

	// populated, err := store.populateLoop(loop, start.PlayerID)
	// if err != nil {
	// 	logger.Println("populate err:", err)
	// } else {
	// 	logger.Println("populated:")
	// 	for _, crd := range populated {
	// 		logger.Println(crd)
	// 	}
	// }

	insideCells, err := findCellsInsideLoop(loop)
	if err != nil {
		logger.Println("failed to find inside cells", err)
		return false
	}

	for _, crd := range insideCells {
		cell := store.field.Get(crd)
		if cell == nil || cell.CellType != cmodels.CellTypeWall || cell.PlayerID != start.PlayerID {
			logger.Println("inside cell:", crd)

			var event models.Event = models.NewSetWallEvent(
				crd,
				start.PlayerID)
			store.Handle(&event)
		}
	}

	return true
}

func (store *Store) startRemovingImmortals() {
	f, err := os.OpenFile("logs/immortals.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	logger = log.New(f, "", os.O_RDWR|os.O_CREATE|os.O_APPEND)

	for {
		if store.components.Len() == 0 {
			logger.Println("no components")
			time.Sleep(BIG_TIMEOUT)
			continue
		}

		bigComponents := make([]*cmodels.Component, 0)
		store.components.Range(func(key uint, cmp *cmodels.Component) bool {
			// logger.Println("wall count", cID, len(cmp.Walls))
			if len(cmp.Walls) >= 10 {
				bigComponents = append(bigComponents, cmp)
			}
			return true
		})
		if len(bigComponents) == 0 {
			logger.Println("no big components found")
			time.Sleep(BIG_TIMEOUT)
			continue
		}

		immortalsWereFound := false
		for _, component := range bigComponents {
			f := processComponent(component)
			immortalsWereFound = immortalsWereFound || f
			logger.Println("immortals found: ", component.ID, f)
		}

		if immortalsWereFound {
			time.Sleep(SMALL_TIMEOUT)
		} else {
			time.Sleep(BIG_TIMEOUT)
		}
	}
}

func exists(list []*cmodels.Cell, element *cmodels.Cell) bool {
	for _, el := range list {
		if el == element {
			return true
		}
	}

	return false
}
func existsCrd(list []cmodels.Coordinates, element cmodels.Coordinates) bool {
	for _, el := range list {
		if el == element {
			return true
		}
	}

	return false
}

func except(list0 []*cmodels.Cell, list1 []*cmodels.Cell) []*cmodels.Cell {
	result := make([]*cmodels.Cell, 0)

	for _, el := range list0 {
		if !exists(list1, el) {
			result = append(result, el)
		}
	}

	return result
}

func (store *Store) findWallLoop(playerID uuid.UUID, startCell *cmodels.Cell) ([]*cmodels.Cell, bool) {
	return findWallLoopRec(store, playerID, startCell, nil, make([]*cmodels.Cell, 0), make([]*cmodels.Cell, 0))
}

func findWallLoopRec(
	store *Store,
	playerID uuid.UUID,
	currentCell *cmodels.Cell,
	previousCell *cmodels.Cell,
	currentPath []*cmodels.Cell,
	visitedCells []*cmodels.Cell,
) ([]*cmodels.Cell, bool) {

	logger.Println("findWallLoopRec: cell:", currentCell.Crd)
	newPath := append(currentPath, currentCell)

	wallNeghbours := store.field.GetOwnWallNeibhours(currentCell.Crd, playerID)
	// End of a wall line. Loop is not found.
	if len(wallNeghbours) == 0 {
		return currentPath, false
	}

	logger.Println("wallNeghbours:")
	var foundLoopEnd *cmodels.Cell = nil
	for _, n := range wallNeghbours {
		log.Print(n.Crd)
		for _, c := range currentPath {
			if n != previousCell && c != previousCell && n == c {
				if len(currentPath) > 1 && n != currentPath[len(currentPath)-2] {
					foundLoopEnd = n
				}
			}
		}
	}
	if foundLoopEnd != nil {
		logger.Println("found loop")
		pathWithoutHead := make([]*cmodels.Cell, 0)
		startFound := false
		for _, cell := range newPath {
			startFound = startFound || foundLoopEnd == cell
			if startFound {
				pathWithoutHead = append(pathWithoutHead, cell)
			}
		}
		return append(pathWithoutHead, foundLoopEnd), true
	}

	filteredWallNeighbours := make([]*cmodels.Cell, 0)
	for _, n := range wallNeghbours {
		wallNeighboursOfNeighbour := store.field.GetOwnWallNeibhours(n.Crd, playerID)

		// We are interested only in walls not surrounded by other own walls.
		// And which have other neighbour walls except previous ones.
		if len(wallNeighboursOfNeighbour) < 6 && len(wallNeighboursOfNeighbour) > 1 {
			if !exists(visitedCells, n) {
				filteredWallNeighbours = append(filteredWallNeighbours, n)
			}
		}
	}

	// Nos suitable wall line ontinuation found. Loop is not found.
	if len(filteredWallNeighbours) == 0 {
		return currentPath, false
	}

	var currentAngle float64
	if previousCell == nil {
		currentAngle = 0
	} else {
		currentAngle = getAngle(currentCell.Crd, previousCell.Crd)
	}
	logger.Println("currentAngle:", currentAngle)

	// We are going to get wall with angle closest to the one we came.
	// So we need to sort walls.
	sort.SliceStable(filteredWallNeighbours, func(i, j int) bool {
		firstAngle := getAngle(currentCell.Crd, filteredWallNeighbours[i].Crd)
		secondAngle := getAngle(currentCell.Crd, filteredWallNeighbours[j].Crd)
		return math.Mod(360+firstAngle-currentAngle, 360) < math.Mod(360+secondAngle-currentAngle, 360)
	})

	logger.Println("filteredWallNeighbours:")
	for _, n := range filteredWallNeighbours {
		angle := getAngle(currentCell.Crd, n.Crd)
		diff0 := 360 + angle - currentAngle
		diff := math.Mod(diff0, 360)
		log.Print(n.Crd, angle, currentAngle, diff0, diff)
	}

	newVisitedCells := append(visitedCells, currentCell)

	for _, n := range filteredWallNeighbours {
		result, found := findWallLoopRec(store, playerID, n, currentCell, newPath, newVisitedCells)
		if found {
			return result, true
		}
	}

	return newPath, false
}
