package main

import (
	"fmt"
	"github.com/RomainMichau/velib_analyzer_go/clients/api"
	"github.com/RomainMichau/velib_analyzer_go/clients/database"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type DbExporter struct {
	api                    *api.VelibApiClient
	database               database.IDatabase
	ticker                 <-chan time.Time
	insertStationCount     int
	insertVelibCount       int
	insertVelibDockedCount int
	runId                  int
	wg                     sync.WaitGroup
	lastStationForVelib    map[int]int
	minorErrorCount        int
	majorErrorCount        int
}

func InitDbExporter(api *api.VelibApiClient, sql database.IDatabase, workerNb int,
	requestMaxFreqMs time.Duration) *DbExporter {
	exporter := &DbExporter{
		api:      api,
		database: sql,
		ticker:   time.Tick(requestMaxFreqMs * time.Millisecond),
	}
	for i := 0; i < workerNb; i++ {
		go exporter.worker()
	}
	return exporter
}

func (exp *DbExporter) RunExport() error {
	exp.insertVelibDockedCount = 0
	exp.minorErrorCount = 0
	exp.insertVelibCount = 0
	exp.insertStationCount = 0
	start := time.Now()
	runId, err := exp.database.InsertRun(time.Now())
	log.Infof("RunID: %d", runId)
	exp.runId = runId
	if err != nil {
		exp.majorErrorCount++
		return fmt.Errorf("failed to insert run in SQL: %w", err)
	}

	allStations, err := exp.api.GetAllStations()
	if err != nil {
		exp.majorErrorCount++
		return fmt.Errorf("failed to all Station for all velib from SQL: %w", err)
	}
	startLastSt := time.Now()
	exp.lastStationForVelib, err = exp.database.GetLastStationForAllVelib()
	elapsedLastSt := time.Since(startLastSt)
	log.Infof("Took %s to get List of Last Station for All velib. ", elapsedLastSt)
	if err != nil {
		exp.majorErrorCount++
		return fmt.Errorf("failed to get Last Statin for all velib from database: %w", err)
	}
	for _, v := range allStations {
		exp.wg.Add(1)
		<-exp.ticker
		err := exp.api.QueueGetVelibRequest(v.Name)
		if err != nil {
			exp.majorErrorCount++
			return err
		}
	}
	elapsed := time.Since(start)
	exp.wg.Wait()
	err = exp.database.RegisterSuccess(runId, exp.minorErrorCount)
	if err != nil {
		return fmt.Errorf("fail to register success: %w", err)
	}
	err = exp.database.PostSync()
	if err != nil {
		return fmt.Errorf("[RunExport] fail to run post sync: %w", err)
	}
	log.Infof("Run %d time took %s. Inserted Station: %d, Inserted Velib: %d, Insert Docked Velib: %d. Minor issues count: %d",
		runId, elapsed, exp.insertStationCount, exp.insertVelibCount, exp.insertVelibDockedCount, exp.minorErrorCount)
	return nil
}

func (exp *DbExporter) worker() {
	for res := range exp.api.RespChan {
		now := time.Now()
		stationDetail, err := exp.api.ParseGetStationDetailResponse(res)
		if err != nil {
			log.Errorf("Failed to read station detail. %s", err.Error())
			exp.minorErrorCount++
			exp.wg.Done()
			return
		}
		stationCode, err := strconv.Atoi(stationDetail.Station.Code)
		if err != nil {
			log.Errorf("Failed to convert station code (%s) to int %s", stationDetail.Station.Code, err.Error())
			exp.wg.Done()
			exp.minorErrorCount++
			return
		}
		stationSql, err := exp.database.GetStationByCode(stationCode)
		if err != nil {
			log.Errorf("Failed to convert station code (%s) to int %s", stationDetail.Station.Code, err.Error())
			exp.wg.Done()
			exp.minorErrorCount++
			return
		}
		if stationSql == nil {
			log.Debugf("Inserting station %s", stationDetail.Station.Name)
			err := exp.database.InsertStation(stationDetail.Station.Name, stationDetail.Station.Gps.Longitude,
				stationDetail.Station.Gps.Latitude, stationCode, exp.runId)
			if err != nil {
				log.Errorf("Failed to insert station %s. %s", stationDetail.Station.Name, err.Error())
				exp.wg.Done()
				exp.minorErrorCount++
				return
			}
			exp.insertStationCount++
		}
		for _, bike := range stationDetail.Bikes {
			velibCode, _ := strconv.Atoi(bike.BikeName)
			sqlBike, err := exp.database.GetVelibByCode(velibCode)
			if err != nil {
				log.Errorf("Failed to get velib by code in SQL. (Code: %d): %s", velibCode, err.Error())
				exp.wg.Done()
				exp.minorErrorCount++
				return
			}
			if sqlBike == nil {
				err := exp.database.InsertVelib(velibCode, exp.runId, bike.BikeElectric == "yes")
				if err != nil {
					log.Errorf("Failed to insert velib %d in SQL: %s", velibCode, err.Error())
					exp.wg.Done()
					exp.minorErrorCount++
					return
				}
				exp.insertVelibCount++
				log.Debugf("Inserting velib %d", velibCode)
			}
			lastStation, present := exp.lastStationForVelib[velibCode]
			if !present || lastStation != stationCode {
				log.Debug("Inserting velib docked ", velibCode, stationCode)
				err := exp.database.InsertVelibDocked(velibCode, stationCode, exp.runId, now, bike.BikeStatus == "disponible")
				if err != nil {
					log.Errorf("Failed to insert velib docked %d in SQL: %s", velibCode, err.Error())
					exp.wg.Done()
					exp.minorErrorCount++
					return
				}
				exp.insertVelibDockedCount++
			}
		}
		exp.wg.Done()
	}
}
