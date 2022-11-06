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

type VelibDockedSql struct {
	VelibCode   int
	StationCode int
	Available   bool
	timeStamp   time.Time
}
