package main

import (
	"encoding/json"
	"fmt"
	"github.com/RomainMichau/velib_analyzer_go/clients/database"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

//48.834882358514875, 2.3045250711792886

type Controller struct {
	sql     *database.VelibSqlClient
	router  *mux.Router
	metrics *Metrics
}

func InitController(sql *database.VelibSqlClient, metrics *Metrics) *Controller {
	r := mux.NewRouter()
	controller := Controller{
		sql:     sql,
		router:  r,
		metrics: metrics,
	}
	spa := spaHandler{staticPath: "webapp/dist/velib_analyzer", indexPath: "index.html"}
	r.HandleFunc("/api/last_station/{code}", controller.getVelib).
		Methods("GET")
	r.HandleFunc("/api/get_arrival/{code}", controller.getVelibArrival).
		Methods("GET")
	r.HandleFunc("/api/healthcheck", controller.healthCheck).
		Methods("GET")
	r.HandleFunc("/api/by_dist", controller.getVelibByDist).
		Queries("long", "{long}", "lat", "{lat}", "dist", "{dist}").
		Methods("GET")
	r.PathPrefix("/").Handler(spa)

	return &controller
}

func (c *Controller) Run(port int) {
	loggedRouter := handlers.LoggingHandler(os.Stdout, c.router)
	log.Infof("Starting controller on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), loggedRouter))
}

func (c *Controller) getVelibByDist(w http.ResponseWriter, r *http.Request) {
	log.Infof("Call received mate1")
	vars := mux.Vars(r)
	longSt, presentlo := vars["long"]
	long, err := strconv.ParseFloat(longSt, 32)
	latSt, presentla := vars["lat"]
	lat, err := strconv.ParseFloat(latSt, 32)
	distSt, presentDist := vars["dist"]
	dist := 500
	if presentDist {
		dist, err = strconv.Atoi(distSt)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if !presentlo || !presentla || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := c.sql.GetVelibByMaxDist(dist, long, lat)
	if err != nil {
		log.Errorf("[getVelibByDist] Error when querying sql: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Errorf(fmt.Sprintf("ERROR: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (c *Controller) getVelibArrival(w http.ResponseWriter, r *http.Request) {
	log.Infof("Call received mate1")
	vars := mux.Vars(r)
	code, present := vars["code"]
	if !present {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	codeI, err := strconv.Atoi(code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	res, err := c.sql.GetVelibArrivalPerStation(codeI)
	if err != nil {
		log.Errorf("[getVelibArrival] Error when querying sql: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (c *Controller) getVelib(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code, present := vars["code"]
	if !present {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	codeI, err := strconv.Atoi(code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	res, _ := c.sql.GetAllStationForVelib(codeI)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (c *Controller) healthCheck(w http.ResponseWriter, r *http.Request) {
	failure := c.metrics.getFailure()
	if failure > 0 {
		log.Warnf("Heathcheck fail. Failure count: %d. Returning 500", failure)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("HC fail: %d", failure)))
		return
	}
	log.Debugf("Heathcheck ok")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("HC ok"))
	return
}
