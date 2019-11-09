package main

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"log"
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

func fetchBusesLastPosition() ([]BusPosition, error) {
	var buses []BusPosition

	for i := 0; ; i++ {
		r, c, err := golemioHttpCall(golemioCurrentPositionEndpoint, busesPerPage, i*busesPerPage)

		if c < 400 && err == nil {
			data, err := ioutil.ReadAll(r.Body)
			processFatalError(err)

			var pr PositionResponse
			err = json.Unmarshal(data, &pr)
			processFatalError(err)

			buses = append(buses, pr.Buses...)

			if len(pr.Buses) < busesPerPage {
				break
			}
		} else {
			return []BusPosition{}, err
		}
	}

	return buses, nil
}

func refreshBusesLastPosition() []BusPosition {
	buses, err := fetchBusesLastPosition()

	if err != nil {
		log.Println("WARNING! Error when fetching last buses position:", err)
	}

	go func() {
		for _, bus := range buses {
			go func(tripID string) {
				_, _ = getTrip(tripID)
			}(bus.Properties.Trip.TripID)
		}
	}()

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
	numBuses := len(buses)
	info := make([]BusInfo, 0, numBuses)

	type tripResult struct {
		i    int
		trip *GolemioTrip
	}

	tripChan := make(chan tripResult, numBuses)
	var wg sync.WaitGroup
	wg.Add(numBuses)

	for i, bus := range buses {
		go func(i int, tripID string) {
			trip, err := getTrip(tripID)
			if err == nil {
				tripChan <- tripResult{i, trip}
			} else {
				log.Printf("WARNING! Error when getting trip \"%s\": %v\n", tripID, err)
			}
			wg.Done()
		}(i, bus.Properties.Trip.TripID)
	}

	go func() {
		wg.Wait()
		close(tripChan)
	}()

	trips := make([]*GolemioTrip, numBuses)

	for trip := range tripChan {
		trips[trip.i] = trip.trip
	}

	for i, bus := range buses {
		trip := trips[i]

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
