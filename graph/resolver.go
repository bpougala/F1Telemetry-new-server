package graph

import (
	"F1Telemetry-new-server/graph/model"
	"sync"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	CurrentLapTimes            []*model.LapTime
	CurrentPositions           []*model.Position
	CurrentCarData             *model.CarData
	CurrentRaceControlMessages []*model.RaceControl
	CurrentStints              []*model.Stint
	CurrentSectorTimes         []*model.Sector
	LapTimeObservers           map[string]chan []*model.LapTime
	PositionObservers          map[string]chan []*model.Position
	CarDataObservers           map[string]chan *model.CarData
	RaceControlObservers       map[string]chan []*model.RaceControl
	StintObservers             map[string]chan []*model.Stint
	SectorTimeObservers        map[string]chan []*model.Sector
	mu                         sync.Mutex
}

func NewResolver() *Resolver {
	return &Resolver{
		CurrentLapTimes:            nil,
		CurrentPositions:           nil,
		CurrentCarData:             nil,
		CurrentRaceControlMessages: nil,
		CurrentStints:              nil,
		CurrentSectorTimes:         nil,
		LapTimeObservers:           make(map[string]chan []*model.LapTime),
		PositionObservers:          make(map[string]chan []*model.Position),
		CarDataObservers:           make(map[string]chan *model.CarData),
		RaceControlObservers:       make(map[string]chan []*model.RaceControl),
		StintObservers:             make(map[string]chan []*model.Stint),
		SectorTimeObservers:        make(map[string]chan []*model.Sector),
	}
}

func (r *Resolver) NotifyLapTimeSubscribers(lapTimes []*model.LapTime) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentLapTimes = append(r.CurrentLapTimes, lapTimes...)
	for _, observer := range r.LapTimeObservers {
		observer <- lapTimes
	}
}

func (r *Resolver) NotifyPositionSubscribers(positions []*model.Position) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentPositions = append(r.CurrentPositions, positions...)
	for _, observer := range r.PositionObservers {
		observer <- positions
	}
}

func (r *Resolver) NotifyCarDataSubscribers(carData *model.CarData) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentCarData = carData
	for _, observer := range r.CarDataObservers {
		observer <- carData
	}
}

func (r *Resolver) NotifyRaceControlSubscribers(raceControlMessages []*model.RaceControl) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentRaceControlMessages = append(r.CurrentRaceControlMessages, raceControlMessages...)
	for _, observer := range r.RaceControlObservers {
		observer <- raceControlMessages
	}
}

func (r *Resolver) NotifyStintSubscribers(stints []*model.Stint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentStints = append(r.CurrentStints, stints...)
	for _, observer := range r.StintObservers {
		observer <- stints
	}
}

func (r *Resolver) NotifySectorTimeSubscribers(sectorTimes []*model.Sector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentSectorTimes = append(r.CurrentSectorTimes, sectorTimes...)
	for _, observer := range r.SectorTimeObservers {
		observer <- sectorTimes
	}
}
