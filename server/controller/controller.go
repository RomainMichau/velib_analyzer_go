package controller

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/RomainMichau/velib_analyzer_go/clients/database"
	"github.com/RomainMichau/velib_analyzer_go/metrics"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

type Controller struct {
	sql     *database.VelibSqlClient
	router  *mux.Router
	metrics *metrics.Metrics
}

//go:embed spec/swagger.json
var openApiSpecs string

func InitController(sql *database.VelibSqlClient, metrics *metrics.Metrics) *Controller {
	r := mux.NewRouter()
	controller := Controller{
		sql:     sql,
		router:  r,
		metrics: metrics,
	}
	spa := spaHandler{staticPath: "webapp/dist/velib_analyzer", indexPath: "index.html"}
	r.HandleFunc("/swagger.json", specHandler).Methods("GET")
	r.HandleFunc("/api/last_station/{code}", controller.GetDockHistory).
		Methods("GET")
	r.HandleFunc("/api/get_arrival/{code}", controller.getVelibArrival).
		Methods("GET")
	r.HandleFunc("/api/healthcheck", controller.healthCheck).
		Methods("GET")
	r.HandleFunc("/api/by_dist", controller.getStationsByDist).
		Queries("long", "{long}", "lat", "{lat}", "dist", "{dist}", "dow", "{dow}").
		Methods("GET")
	r.PathPrefix("/").Handler(spa)

	return &controller
}

func (c *Controller) RunWithTls(port int, certPath string, keyPath string) {
	loggedRouter := handlers.LoggingHandler(os.Stdout, c.router)
	log.Infof("Starting controller on port %d", port)
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%d", port), certPath, keyPath, loggedRouter))
}

func (c *Controller) Run(port int) {
	loggedRouter := handlers.LoggingHandler(os.Stdout, c.router)
	log.Infof("Starting controller on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), loggedRouter))
}

func specHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, openApiSpecs)
}

// GetStationByDist godoc
//
//	@Summary		Return all stations in a certain distance from GPS coordinates
//	@Description	Return all stations in a certain distance from GPS coordinates*
//	@Tags			Velibs
//	@Accept			json
//	@Produce		json
//	@Param			long	query	float32						true	"Longitude"						default(2.3391411244733233)
//	@Param			lat		query	float32						true	"Latitude"						default(48.84641747361601)
//	@Param			dist	query	int							true	"Max distance"					default(1000)
//	@Param			dow		query	int							true	"Day of week (mon:1, sun: 7)"	default(1)
//	@Success		200		{array}	clients.StationWithArrivals	"List of velib arrivals for requested station"
//	@Failure		400		"Invalid params"
//	@Router			/api/by_dist [get]
func (c *Controller) getStationsByDist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	longSt, presentlo := vars["long"]
	long, err := strconv.ParseFloat(longSt, 32)
	latSt, presentla := vars["lat"]
	lat, err := strconv.ParseFloat(latSt, 32)
	distSt, presentDist := vars["dist"]
	dowSt, presentdow := vars["dow"]
	dow, err := strconv.Atoi(dowSt)
	dist := 500
	if presentDist {
		dist, err = strconv.Atoi(distSt)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if !presentlo || !presentla || !presentdow || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := c.sql.GetVelibByMaxDistAndArrival(dist, long, lat, dow)
	if err != nil {
		log.Errorf("[getStationsByDist] Error when querying sql: %s", err.Error())
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

// GetVelibArrival godoc
//
//	@Summary		Return avg velib arrivals
//	@Description	Return avg velib arrivals per dow and how for a requested station
//	@Tags			Velibs
//	@Accept			json
//	@Produce		json
//	@Param			code	path	int						true	"Station code"	default(15122)
//	@Success		200		{array}	clients.VelibArrival	"List of velib arrivals for requested station"
//	@Failure		400		"Invalid params"
//	@Router			/api/get_arrival/{code} [get]
func (c *Controller) getVelibArrival(w http.ResponseWriter, r *http.Request) {
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

// GetDockHistory godoc
//
//	@Summary		Return dock history for velib
//	@Description	Return dock history for velib
//	@Tags			Velibs
//	@Accept			json
//	@Produce		json
//	@Param			code	path	int								true	"velib code"	default(60549)
//	@Success		200		{array}	clients.VelibDockedSqlDetails	"List of velib arrivals for requested station"
//	@Router			/api/last_station/{code} [get]
func (c *Controller) GetDockHistory(w http.ResponseWriter, r *http.Request) {
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
	failure := c.metrics.GetFailure()
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
