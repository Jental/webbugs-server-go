package main

import (
	"time"
)

var interval2 time.Duration = 1 * time.Minute

// StartPlayerRemover - starts inactive player remove process
func (store *Store) StartPlayerRemover() {
	ticker := time.NewTicker(interval)

	go func() {
		for {
			<-ticker.C

			// now := time.Now().UTC()
		}
	}()
}
