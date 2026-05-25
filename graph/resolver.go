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
	CurrentSegments            []*model.Segment
	CurrentTrackStatus         []*model.TrackStatus
	CurrentWeather             *model.Weather
	CurrentDriverLocations     *model.DriverLocation
	LapTimeObservers           map[string]chan []*model.LapTime
	PositionObservers          map[string]chan []*model.Position
	CarDataObservers           map[string]chan *model.CarData
	RaceControlObservers       map[string]chan []*model.RaceControl
	StintObservers             map[string]chan []*model.Stint
	SectorTimeObservers        map[string]chan []*model.Sector
	SegmentObservers           map[string]chan []*model.Segment
	TrackStatusObservers       map[string]chan []*model.TrackStatus
	WeatherObservers           map[string]chan *model.Weather
	DriverLocationObservers    map[string]chan *model.DriverLocation
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
		CurrentTrackStatus:         nil,
		LapTimeObservers:           make(map[string]chan []*model.LapTime),
		PositionObservers:          make(map[string]chan []*model.Position),
		CarDataObservers:           make(map[string]chan *model.CarData),
		RaceControlObservers:       make(map[string]chan []*model.RaceControl),
		StintObservers:             make(map[string]chan []*model.Stint),
		SectorTimeObservers:        make(map[string]chan []*model.Sector),
		SegmentObservers:           make(map[string]chan []*model.Segment),
		TrackStatusObservers:       make(map[string]chan []*model.TrackStatus),
		WeatherObservers:           make(map[string]chan *model.Weather),
		DriverLocationObservers:    make(map[string]chan *model.DriverLocation),
	}
}

func (r *Resolver) NotifyLapTimeSubscribers(lapTimes []*model.LapTime) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentLapTimes = append(r.CurrentLapTimes, lapTimes...)
	for _, observer := range r.LapTimeObservers {
		select {
		case observer <- lapTimes:
		default:
		}
	}
}

func (r *Resolver) NotifyPositionSubscribers(positions []*model.Position) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentPositions = append(r.CurrentPositions, positions...)
	for _, observer := range r.PositionObservers {
		select {
		case observer <- positions:
		default:
		}
	}
}

func (r *Resolver) NotifyCarDataSubscribers(carData *model.CarData) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentCarData = carData
	for _, observer := range r.CarDataObservers {
		select {
		case observer <- carData:
		default:
		}
	}
}

func (r *Resolver) NotifyRaceControlSubscribers(raceControlMessages []*model.RaceControl) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentRaceControlMessages = append(r.CurrentRaceControlMessages, raceControlMessages...)
	for _, observer := range r.RaceControlObservers {
		select {
		case observer <- raceControlMessages:
		default:
		}
	}
}

func (r *Resolver) NotifyStintSubscribers(stints []*model.Stint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentStints = append(r.CurrentStints, stints...)
	for _, observer := range r.StintObservers {
		select {
		case observer <- stints:
		default:
		}
	}
}

func (r *Resolver) NotifySectorTimeSubscribers(sectorTimes []*model.Sector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentSectorTimes = append(r.CurrentSectorTimes, sectorTimes...)
	for _, observer := range r.SectorTimeObservers {
		select {
		case observer <- sectorTimes:
		default:
		}
	}
}

func (r *Resolver) NotifySegmentSubscribers(segments []*model.Segment) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentSegments = segments
	for _, observer := range r.SegmentObservers {
		select {
		case observer <- segments:
		default:
		}
	}
}

func (r *Resolver) NotifyTrackStatusSubscribers(trackStatuses []*model.TrackStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentTrackStatus = append(r.CurrentTrackStatus, trackStatuses...)
	for _, observer := range r.TrackStatusObservers {
		select {
		case observer <- trackStatuses:
		default:
		}
	}
}

func (r *Resolver) NotifyWeatherSubscribers(weather *model.Weather) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentWeather = weather
	for _, observer := range r.WeatherObservers {
		select {
		case observer <- weather:
		default:
		}
	}
}

func (r *Resolver) NotifyDriverLocationSubscribers(location *model.DriverLocation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentDriverLocations = location
	for _, observer := range r.DriverLocationObservers {
		select {
		case observer <- location:
		default:
		}
	}
}

func (r *Resolver) RegisterDriverLocationObserver(id string, ch chan *model.DriverLocation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.DriverLocationObservers[id] = ch
}

func (r *Resolver) RegisterWeatherObserver(id string, ch chan *model.Weather) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.WeatherObservers[id] = ch
}

func (r *Resolver) RegisterLapTimeObserver(id string, ch chan []*model.LapTime) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.LapTimeObservers[id] = ch
}

func (r *Resolver) RegisterPositionObserver(id string, ch chan []*model.Position) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.PositionObservers[id] = ch
}

func (r *Resolver) RegisterCarDataObserver(id string, ch chan *model.CarData) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CarDataObservers[id] = ch
}

func (r *Resolver) RegisterRaceControlObserver(id string, ch chan []*model.RaceControl) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RaceControlObservers[id] = ch
}

func (r *Resolver) RegisterStintObserver(id string, ch chan []*model.Stint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.StintObservers[id] = ch
}

func (r *Resolver) RegisterSectorTimeObserver(id string, ch chan []*model.Sector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.SectorTimeObservers[id] = ch
}

func (r *Resolver) RegisterSegmentObserver(id string, ch chan []*model.Segment) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.SegmentObservers[id] = ch
}

func (r *Resolver) RegisterTrackStatusObserver(id string, ch chan []*model.TrackStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.TrackStatusObservers[id] = ch
}
