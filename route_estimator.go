package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
)

type (
	position struct {
		latitude  float32
		longitude float32
	}

	RouteAPIResponse struct {
		ResourceSets []struct {
			Resources []struct {
				Duration int `json:"travelDurationTraffic"`
			} `json:"resources"`
		} `json:"resourceSets"`
	}

	Estimate struct {
		Estimate float32 `json:"estimate"`
		BusID    string  `json:"bus_id"`
		StopID   string  `json:"stop_id"`
	}

	BusNotFoundError         struct{}
	StopNotInPathError       struct{}
	UnspecifiedEstimateError struct{}
)

var (
	routingEndpoint = "http://dev.virtualearth.net/REST/V1/Routes/Driving?key="
)

func (e *BusNotFoundError) Error() string {
	return "target bus not found"
}

func (e *StopNotInPathError) Error() string {
	return "stop not in bus path"
}

func (e *UnspecifiedEstimateError) Error() string {
	return "unspecified estimate error"
}

func positionToWpString(p *position, i int) string {
	return fmt.Sprintf("&wp.%d=%.6f,%.6f", i, p.latitude, p.longitude)
}

func estimateTimeToStop(busID string, targetStopID string) (Estimate, error) {
	rtn := Estimate{0, busID, targetStopID}

	buses := getCurrentBusInfo()
	var targetBus *BusInfo

	for _, bus := range buses {
		if bus.ID == busID {
			targetBus = &bus
			break
		}
	}

	if targetBus == nil {
		return rtn, new(BusNotFoundError)
	}

	lat, _ := strconv.ParseFloat(targetBus.Latitude, 32)
	lon, _ := strconv.ParseFloat(targetBus.Longitude, 32)

	start := position{float32(lat), float32(lon)}
	stops := targetBus.Stops
	nextStopID := targetBus.NextStopID

	routeString := positionToWpString(&start, 0)
	var routeStrings []string
	writeWp := false
	targetStopFound := false
	routeI := 1

	for _, stop := range stops {
		if !writeWp && nextStopID == stop.ID {
			writeWp = true
		}

		if writeWp {
			routeString += positionToWpString(&position{stop.Latitude, stop.Longitude}, routeI)
			routeI++
		}

		if routeI == 24 {
			routeI = 0
			routeStrings = append(routeStrings, routeString)
			routeString = ""
		}

		if writeWp && stop.ID == targetStopID {
			targetStopFound = true
			break
		}
	}

	if !targetStopFound {
		return rtn, new(StopNotInPathError)
	}

	if len(routeString) > 0 {
		routeStrings = append(routeStrings, routeString)
	}

	if len(routeStrings) == 0 {
		return rtn, new(UnspecifiedEstimateError)
	}

	estimate := float32(0)

	for _, str := range routeStrings {
		resp, _ := httpCall("GET", routingEndpoint+str, nil)
		data, err := ioutil.ReadAll(resp.Body)
		processFatalError(err)

		var result RouteAPIResponse
		err = json.Unmarshal(data, &result)
		processFatalError(err)

		if len(result.ResourceSets) > 0 && len(result.ResourceSets[0].Resources) > 0 {
			estimate += float32(result.ResourceSets[0].Resources[0].Duration) / 60
		}
	}

	rtn.Estimate = estimate

	return rtn, nil
}

func setMicrosoftApiKey(key string) {
	routingEndpoint += key
}
