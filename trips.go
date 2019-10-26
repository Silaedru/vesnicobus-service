package main

import (
	"encoding/json"
	"io/ioutil"
)

type GolemioStopProperties struct {
	Id string `json:"stop_id"`
	Name string `json:"stop_name"`
	Latitude float32 `json:"stop_lat"`
	Longitude float32 `json:"stop_lon"`
}

type GolemioStop struct {
	Properties GolemioStopProperties `json:"properties"`
}

type GolemioTrip struct {
	TripId string `json:"trip_id"`
	Stops []GolemioStop `json:"stops"`
}

const (
	golemioTripEndpoint = "https://api.golemio.cz/v1/gtfs/trips/"
)

func tripKey(id string) string {
	return "trip_"+id
}

func loadTrip(id string) *GolemioTrip {
	trip, rc := golemioHttpCall(golemioTripEndpoint + id + "?includeStops=true", 1, 0)

	if rc < 400 {
		data, err := ioutil.ReadAll(trip.Body)
		processFatalError(err)

		var tripData GolemioTrip
		err = json.Unmarshal(data, &tripData)
		processFatalError(err)

		return &tripData
	}

	return nil
}

func getTrip(id string) *GolemioTrip {
	tripData := getItem(tripKey(id))
	var trip *GolemioTrip

	if tripData == "" {
		trip = loadTrip(id)

		if trip != nil {
			tripData, err := json.Marshal(trip)
			processFatalError(err)
			storeItem(tripKey(id), string(tripData))
		}
	} else {
		err := json.Unmarshal([]byte(tripData), &trip)
		processFatalError(err)
	}

	return trip
}