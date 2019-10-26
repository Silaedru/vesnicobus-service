package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type position struct {
	latitude float32
	longitude float32
}

type RouteAPIResponse struct {
	ResourceSets []struct {
		Resources []struct {
			Duration int `json:"travelDurationTraffic"`
		} `json:"resources"`
	} `json:"resourceSets"`
}

var routingEndpoint = "http://dev.virtualearth.net/REST/V1/Routes/Driving?key="

func positionToWpString(p *position, i int) string {
	return fmt.Sprintf("&wp.%d=%.6f,%.6f", i, p.latitude, p.longitude)
}

func estimateTimeToStop(start position, nextStopId string, stops *[]GolemioStopProperties, targetStopId string) (float32, error) {
	routeString := positionToWpString(&start, 0)
	var routeStrings []string
	writeWp := false
	targetStopFound := false
	routeI := 1

	for _, stop := range *stops {
		if !writeWp && nextStopId == stop.Id {
			writeWp = true
		}

		if writeWp {
			routeString += positionToWpString(&position{stop.Latitude, stop.Longitude}, routeI)
			routeI++
		}

		if routeI == 24  {
			routeI = 0
			routeStrings = append(routeStrings, routeString)
			routeString = ""
		}

		if writeWp && stop.Id == targetStopId {
			targetStopFound = true
			break
		}
	}

	if !targetStopFound {
		return 0, fmt.Errorf("target stop not in bus path")
	}

	if len(routeString) > 0 {
		routeStrings = append(routeStrings, routeString)
	}

	if len(routeStrings) == 0 {
		return 0, fmt.Errorf("unspecified estimate error")
	}

	estimate := float32(0)

	for _, str := range routeStrings {
		resp, _ := httpCall("GET", routingEndpoint+str, nil)
		data, err := ioutil.ReadAll(resp.Body)
		processFatalError(err)
		var result RouteAPIResponse
		err = json.Unmarshal(data, &result)
		processFatalError(err)
		estimate += float32(result.ResourceSets[0].Resources[0].Duration)/60
	}

	return estimate, nil
}

func setMicrosoftApiKey(key string) {
	routingEndpoint += key
}