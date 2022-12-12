package database

import (
	"github.com/RomainMichau/velib_finder/clients"
	"time"
)

type IDatabase interface {
	GetAllVelibCode() []int
	GetLastStationForAllVelib() (map[int]int, error)
	GetVelibByCode(code int) (*clients.VelibSqlEntity, error)
	GetLastDockedForVelib(velibCode int) (*clients.VelibDockedSql, error)
	GetStationByCode(code int) (*clients.StationSqlEntity, error)
	GetAllStationsCode() (map[int]bool, error)
	GetAllStationForVelib(velibCode int) ([]clients.VelibDockedSqlDetails, error)
	InsertVelib(velibCode, run int, electric bool) error
	InsertStation(name string, longitude, latitude float32, code, runId int) error
	InsertVelibDocked(velibCode, stationCode, runId int, time time.Time, available bool) error
	InsertRun(time time.Time) (int, error)
	RegisterSuccess(runId, minorIssueCount int) error
	PostSync() error
}
