package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func setResponseHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// /buses/{bus_id}/estimate/{stop_id}
func handleEstimate(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	busId := params["bus_id"]
	stopId := params["stop_id"]

	buses := getCurrentBusInfo()
	var targetBus *BusInfo

	for _, bus := range buses {
		if bus.Id == busId {
			targetBus = &bus
			break
		}
	}

	if targetBus == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	lat, _ := strconv.ParseFloat(targetBus.Latitude, 32)
	lon, _ := strconv.ParseFloat(targetBus.Longitude, 32)

	estimate, err := estimateTimeToStop(position{float32(lat), float32(lon)}, targetBus.NextStopId, &targetBus.Stops, stopId)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	_, err = w.Write([]byte(fmt.Sprintf("{\"estimate\":%.2f}", estimate)))
	processFatalError(err)
}

// GET /buses
func handleBuses(w http.ResponseWriter, r *http.Request) {
	bi := getCurrentBusInfo()

	resp, err := json.Marshal(bi)
	processFatalError(err)

	_, err = w.Write(resp)
	processFatalError(err)
}

func setupWebService(bindAddr string) {
	router := mux.NewRouter()

	router.HandleFunc("/buses", handleBuses).Methods("GET")
	router.HandleFunc("/buses/{bus_id}/estimate/{stop_id}", handleEstimate).Methods("GET")

	router.Use(setResponseHeaders)

	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "OPTIONS"})

	log.Println("starting webservice ...")
	err := http.ListenAndServe(bindAddr, handlers.CORS(allowedOrigins, allowedHeaders, allowedMethods)(router))
	processFatalError(err)
}