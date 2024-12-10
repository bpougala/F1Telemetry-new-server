package graph

import (
	"F1Telemetry-new-server/graph/model"
	"sync"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	CurrentLapTimes   []*model.LapTime
	CurrentPositions  []*model.Position
	LapTimeObservers  map[string]chan []*model.LapTime
	PositionObservers map[string]chan []*model.Position
	mu                sync.Mutex
}

func NewResolver() *Resolver {
	return &Resolver{
		CurrentLapTimes:   nil,
		CurrentPositions:  nil,
		LapTimeObservers:  make(map[string]chan []*model.LapTime),
		PositionObservers: make(map[string]chan []*model.Position),
	}
}

func (r *Resolver) NotifyLapTimeSubscribers(lapTimes []*model.LapTime) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentLapTimes = lapTimes
	for _, observer := range r.LapTimeObservers {
		observer <- lapTimes
	}
}

func (r *Resolver) NotifyPositionSubscribers(positions []*model.Position) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentPositions = positions
	for _, observer := range r.PositionObservers {
		observer <- positions
	}
}
