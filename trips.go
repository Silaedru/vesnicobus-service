package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type (
	GolemioStopProperties struct {
		ID        string  `json:"stop_id"`
		Name      string  `json:"stop_name"`
		Latitude  float32 `json:"stop_lat"`
		Longitude float32 `json:"stop_lon"`
	}

	GolemioStop struct {
		Properties GolemioStopProperties `json:"properties"`
	}

	GolemioTrip struct {
		TripID string        `json:"trip_id"`
		Stops  []GolemioStop `json:"stops"`
	}
)

const (
	golemioTripEndpoint = "https://api.golemio.cz/v1/gtfs/trips/"
)

var (
	tripLocks        = make(map[string]*sync.Mutex)
	tripLockMapMutex = &sync.Mutex{}
)

func tripKey(id string) string {
	return "trip_" + id
}

func loadTrip(id string) (*GolemioTrip, error) {
	trip, rc, err := golemioHttpCall(golemioTripEndpoint+id+"?includeStops=true", 1, 0)

	if rc < 400 && err == nil {
		data, err := ioutil.ReadAll(trip.Body)
		processFatalError(err)

		var tripData GolemioTrip
		err = json.Unmarshal(data, &tripData)
		processFatalError(err)

		return &tripData, nil
	}

	return nil, err
}

func getTrip(id string) (*GolemioTrip, error) {
	tripLockMapMutex.Lock()
	if tripLocks[id] == nil {
		tripLocks[id] = new(sync.Mutex)
	}
	tripLock := tripLocks[id]
	tripLockMapMutex.Unlock()

	tripLock.Lock()
	tripData := getItem(tripKey(id))
	var trip *GolemioTrip

	if tripData == "" {
		trip, err := loadTrip(id)

		if err != nil {
			return nil, err
		}

		if trip != nil {
			tripData, err := json.Marshal(trip)
			processFatalError(err)
			storeItem(tripKey(id), string(tripData))
		}
		tripLock.Unlock()
	} else {
		tripLock.Unlock()
		err := json.Unmarshal([]byte(tripData), &trip)
		processFatalError(err)
	}

	return trip, nil
}
