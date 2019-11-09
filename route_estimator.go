package main

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

	routeEstimator interface {
		estimateTimeToStop(busID string, targetStopID string) (Estimate, error)
	}

	BusNotFoundError         struct{}
	StopNotInPathError       struct{}
	UnspecifiedEstimateError struct{}
)

var (
	estimator routeEstimator
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
