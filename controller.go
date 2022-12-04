package main

import (
	"encoding/json"
	"fmt"
	"github.com/RomainMichau/velib_finder/clients"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type Controller struct {
	sql    *clients.VelibSqlClient
	router *mux.Router
}

func InitController(sql *clients.VelibSqlClient) *Controller {
	r := mux.NewRouter()
	controller := Controller{
		sql:    sql,
		router: r,
	}

	r.HandleFunc("/last_station/{code}", controller.getVelib).
		Methods("GET")

	return &controller
}

func (c *Controller) Run(port int) {
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), c.router)
	if err != nil {
		panic(err.Error())
	}
}

func (c *Controller) getVelib(w http.ResponseWriter, r *http.Request) {
	log.Infof("Call received")
	vars := mux.Vars(r)
	code, present := vars["code"]
	if !present {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	code_i, err := strconv.Atoi(code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	res, _ := c.sql.GetAllStationForVelib(code_i)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
