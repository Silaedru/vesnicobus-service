package main

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"sync"
)

type (
	BusPosition struct {
		Properties struct {
			Trip struct {
				TripID string `json:"gtfs_trip_id"`
				Line   string `json:"cis_short_name"`
			} `json:"trip"`

			LastPosition struct {
				NextStopID string `json:"gtfs_next_stop_id"`
				Latitude   string `json:"lat"`
				Longitude  string `json:"lng"`
				Delay      int    `json:"delay"`
			} `json:"last_position"`
		} `json:"properties"`
	}

	PositionResponse struct {
		Buses []BusPosition `json:"features"`
	}

	BusInfo struct {
		ID           string                  `json:"id"`
		Line         string                  `json:"line"`
		NextStopName string                  `json:"next_stop_name"`
		NextStopID   string                  `json:"next_stop_id"`
		LastStopName string                  `json:"last_stop_name"`
		Latitude     string                  `json:"latitude"`
		Longitude    string                  `json:"longitude"`
		Stops        []GolemioStopProperties `json:"stops"`
		Delay        float32                 `json:"delay"`
	}

	BusInfoResponse struct {
		BusInfo   []BusInfo `json:"bus_info"`
		Timestamp int64     `json:"timestamp"`
	}
)

const (
	golemioCurrentPositionEndpoint = "https://api.golemio.cz/v1/vehiclepositions?"
	busesPerPage                   = 1000
	lastPositionKey                = "lastposition"
)

var (
	refreshPositionLock = &sync.Mutex{}
)

func fetchBusesLastPosition() []BusPosition {
	var buses []BusPosition

	for i := 0; ; i++ {
		r, c := golemioHttpCall(golemioCurrentPositionEndpoint, busesPerPage, i*busesPerPage)
		if c < 400 {
			data, err := ioutil.ReadAll(r.Body)
			processFatalError(err)

			var pr PositionResponse
			err = json.Unmarshal(data, &pr)
			processFatalError(err)

			buses = append(buses, pr.Buses...)

			if len(pr.Buses) < busesPerPage {
				break
			}
		}
	}

	return buses
}

func refreshBusesLastPosition() []BusPosition {
	buses := fetchBusesLastPosition()
	tripData, err := json.Marshal(buses)
	processFatalError(err)

	storeItem(lastPositionKey, string(tripData))
	return buses
}

func getBusesLastPosition() []BusPosition {
	refreshPositionLock.Lock()
	lastPositionData := getItem(lastPositionKey)
	var buses []BusPosition

	if lastPositionData == "" {
		buses = refreshBusesLastPosition()
		refreshPositionLock.Unlock()
	} else {
		refreshPositionLock.Unlock()
		err := json.Unmarshal([]byte(lastPositionData), &buses)
		processFatalError(err)
	}

	return buses
}

func getCurrentBusInfo() []BusInfo {
	buses := getBusesLastPosition()
	info := make([]BusInfo, 0, len(buses))

	for _, bus := range buses {
		trip := getTrip(bus.Properties.Trip.TripID)

		if trip == nil {
			continue
		}

		nextStop := ""

		for _, stop := range trip.Stops {
			if stop.Properties.ID == bus.Properties.LastPosition.NextStopID {
				nextStop = stop.Properties.Name
				break
			}
		}

		stops := make([]GolemioStopProperties, len(trip.Stops))

		for j, stop := range trip.Stops {
			stops[j] = stop.Properties
		}

		busId := fmt.Sprintf("%x", crc32.ChecksumIEEE([]byte(trip.TripID)))
		lastStop := ""

		if len(trip.Stops) > 1 {
			lastStop = trip.Stops[len(trip.Stops)-1].Properties.Name
		}

		info = append(info, BusInfo{busId,
			bus.Properties.Trip.Line,
			nextStop,
			bus.Properties.LastPosition.NextStopID,
			lastStop,
			bus.Properties.LastPosition.Latitude,
			bus.Properties.LastPosition.Longitude,
			stops,
			float32(bus.Properties.LastPosition.Delay) / 60,
		})
	}

	return info
}
