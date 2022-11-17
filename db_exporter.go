package main

import (
	"github.com/RomainMichau/velib_finder/clients"
	"log"
	"strconv"
	"sync"
	"time"
)

type DbExporter struct {
	api                 *clients.VelibApiClient
	sql                 *clients.VelibSqlClient
	ticker              <-chan time.Time
	i                   int
	runId               int
	mutex               sync.Mutex
	wg                  sync.WaitGroup
	lastStationForVelib map[int]int
}

func InitDbExporter(api *clients.VelibApiClient, sql *clients.VelibSqlClient, workerNb int,
	requestMaxFreqMs time.Duration) *DbExporter {
	exporter := &DbExporter{
		api:    api,
		sql:    sql,
		ticker: time.Tick(requestMaxFreqMs * time.Millisecond),
	}
	for i := 0; i < workerNb; i++ {
		go exporter.worker()
	}
	return exporter
}

func (exp *DbExporter) worker() {
	for res := range exp.api.RespChan {
		now := time.Now()
		<-exp.ticker
		stationDetail, err := exp.api.ParseGetStationDetailResponse(res)
		if err != nil {
			panic(err)
		}
		stationCode, err := strconv.Atoi(stationDetail.Station.Code)
		if err != nil {
			panic(err)
		}
		stationSql, err := exp.sql.GetStationByCode(stationCode)
		if err != nil {
			panic(err)
		}
		if stationSql == nil {
			log.Println("Inserting station ", stationDetail.Station.Name)
			err := exp.sql.InsertStation(stationDetail.Station.Name, stationDetail.Station.Gps.Longitude,
				stationDetail.Station.Gps.Latitude, stationCode, exp.runId)
			if err != nil {
				print(err)
			}
		}
		for _, bike := range stationDetail.Bikes {
			velibCode, _ := strconv.Atoi(bike.BikeName)
			sqlBike, err := exp.sql.GetVelibByCode(velibCode)
			if err != nil {
				panic(err)
			}
			if sqlBike == nil {
				err := exp.sql.InsertVelib(velibCode, exp.runId, bike.BikeElectric == "yes")
				if err != nil {
					return
				}
				log.Println("Inserting velib ", velibCode, exp.i)
				exp.i++
				last_station, present := exp.lastStationForVelib[velibCode]
				if !present || last_station != stationCode {
					log.Println("Inserting velib docked", velibCode, stationCode)
					err := exp.sql.InsertVelibDocked(velibCode, stationCode, exp.runId, now, bike.BikeStatus == "disponible")
					if err != nil {
						panic(err)
					}
				}
			}
		}
		exp.wg.Done()

	}
}

func (exp *DbExporter) RunExport() error {
	start := time.Now()
	runId, err := exp.sql.InsertRun(time.Now())
	exp.runId = runId
	if err != nil {
		return err
	}
	allStations, err := exp.api.GetAllStations()
	if err != nil {
		return err
	}
	exp.lastStationForVelib, err = exp.sql.GetLastStationForAllVelib()
	if err != nil {
		return err
	}
	for _, v := range allStations {
		exp.wg.Add(1)

		err := exp.api.QueueGetVelibRequest(v.Name)
		if err != nil {
			return err
		}
	}
	elapsed := time.Since(start)
	exp.wg.Wait()
	log.Printf("Run time took %s\n", elapsed)
	return nil
}
