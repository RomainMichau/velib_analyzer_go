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

type StationWithArrivals struct {
	Name      string
	Longitude float32
	Latitude  float32
	Code      int
	Dist      float32
	// dow -> hour -> arrival
	Arrival map[int]map[int]float32
}

func (s *StationWithArrivals) AddArrival(dow, hour int, avg float32) {
	if s.Arrival == nil {
		s.Arrival = map[int]map[int]float32{}
	}
	_, dowPres := s.Arrival[dow]
	if !dowPres {
		s.Arrival[dow] = map[int]float32{}
	}
	_, hourPres := s.Arrival[dow][hour]
	if !hourPres {
		s.Arrival[dow][hour] = avg
		return
	}

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
