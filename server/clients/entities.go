package clients

import "time"

type VelibApiEntity struct {
	Name      string
	Longitude float32
	Latitude  float32
	Code      int
}

type VelibSqlEntity struct {
	Id       int
	Code     int
	Electric bool
}

type StationSqlEntity struct {
	Name      string
	Longitude float32
	Latitude  float32
	Code      int
}

type StationSqlEntityWithDist struct {
	Name      string
	Longitude float32
	Latitude  float32
	Code      int
	Dist      float32
}

type VelibDockedSql struct {
	VelibCode   int
	StationCode int
	Available   bool
	TimeStamp   time.Time
}

type VelibDockedSqlDetails struct {
	VelibCode int
	Station   StationSqlEntity
	Available bool
	TimeStamp time.Time
}

type VelibArrival struct {
	Avg  float32
	Dow  int
	Hour int
}
