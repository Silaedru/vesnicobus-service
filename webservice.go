package main

import (
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

func setResponseHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// GET /buses/{bus_id}/estimate/{stop_id}
func handleEstimate(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	busID := params["bus_id"]
	stopID := params["stop_id"]

	estimate, err := estimateTimeToStop(busID, stopID)

	if err != nil {
		switch err.(type) {
		case *BusNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	estimateString, err := json.Marshal(estimate)
	processFatalError(err)

	w.Write(estimateString)
}

// GET /buses
func handleBuses(w http.ResponseWriter, r *http.Request) {
	bi := getCurrentBusInfo()

	bir := BusInfoResponse{bi, time.Now().Unix()}

	resp, err := json.Marshal(bir)
	processFatalError(err)

	w.Write(resp)
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
	err := http.ListenAndServe(bindAddr, handlers.CORS(allowedOrigins, allowedHeaders, allowedMethods)(handlers.CompressHandler(router)))
	processFatalError(err)
}
