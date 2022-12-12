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

type Controller struct {
	sql    *database.VelibSqlClient
	router *mux.Router
}

func InitController(sql *database.VelibSqlClient) *Controller {
	r := mux.NewRouter()
	controller := Controller{
		sql:    sql,
		router: r,
	}

	r.HandleFunc("/last_station/{code}", controller.getVelib).
		Methods("GET")
	r.HandleFunc("/get_arrival/{code}", controller.getVelibArrival).
		Methods("GET")
	return &controller
}

func (c *Controller) Run(port int) {
	loggedRouter := handlers.LoggingHandler(os.Stdout, c.router)
	log.Infof("Starting controller on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), loggedRouter))
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
	log.Infof("Call received mate4")
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
