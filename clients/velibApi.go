package clients

import (
	"encoding/json"
	"fmt"
	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
	"strconv"
	"strings"
	"time"
)

type VelibApiClient struct {
	client cloudscraper.CloudScrapper
}

type stationsApiResponse struct {
	Station struct {
		Gps struct {
			Latitude  float32 `json:"latitude"`
			Longitude float32 `json:"longitude"`
		} `json:"gps"`
		State   string `json:"state"`
		Name    string `json:"name"`
		Code    string `json:"code"`
		Type    string `json:"type"`
		DueDate int    `json:"dueDate"`
	} `json:"station"`
	NbBike             int    `json:"nbBike"`
	NbEbike            int    `json:"nbEbike"`
	NbFreeDock         int    `json:"nbFreeDock"`
	NbFreeEDock        int    `json:"nbFreeEDock"`
	CreditCard         string `json:"creditCard"`
	NbDock             int    `json:"nbDock"`
	NbEDock            int    `json:"nbEDock"`
	NbBikeOverflow     int    `json:"nbBikeOverflow"`
	NbEBikeOverflow    int    `json:"nbEBikeOverflow"`
	KioskState         string `json:"kioskState"`
	Overflow           string `json:"overflow"`
	OverflowActivation string `json:"overflowActivation"`
	MaxBikeOverflow    int    `json:"maxBikeOverflow"`
	DensityLevel       int    `json:"densityLevel"`
}

type stationDetailApiResponse struct {
	Station struct {
		Gps struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"gps"`
		State   string `json:"state"`
		Name    string `json:"name"`
		Code    string `json:"code"`
		Type    string `json:"type"`
		DueDate int    `json:"dueDate"`
	} `json:"station"`
	NbBike             int    `json:"nbBike"`
	NbEbike            int    `json:"nbEbike"`
	NbFreeDock         int    `json:"nbFreeDock"`
	NbFreeEDock        int    `json:"nbFreeEDock"`
	CreditCard         string `json:"creditCard"`
	NbDock             int    `json:"nbDock"`
	NbEDock            int    `json:"nbEDock"`
	NbBikeOverflow     int    `json:"nbBikeOverflow"`
	NbEBikeOverflow    int    `json:"nbEBikeOverflow"`
	KioskState         string `json:"kioskState"`
	Overflow           string `json:"overflow"`
	OverflowActivation string `json:"overflowActivation"`
	MaxBikeOverflow    int    `json:"maxBikeOverflow"`
	DensityLevel       int    `json:"densityLevel"`
	Bikes              []struct {
		DockPosition  string    `json:"dockPosition"`
		BikeName      string    `json:"bikeName"`
		BikeElectric  string    `json:"bikeElectric"`
		BikeStatus    string    `json:"bikeStatus"`
		BikeRate      int       `json:"bikeRate"`
		NumberOfRates int       `json:"numberOfRates"`
		LastRateDate  time.Time `json:"lastRateDate"`
	} `json:"bikes"`
}

type getStationBody struct {
	StationName   string `json:"stationName"`
	Disponibility string `json:"disponibility"`
}

func InitVelibApi() VelibApiClient {
	client, _ := cloudscraper.Init(false)
	api := VelibApiClient{
		client: *client,
	}
	return api
}

var (
	baseUrl                = "https://www.velib-metropole.fr/api"
	allStationsEndpoint    = "/map/details?zoomLevel=1&gpsTopLatitude=49.05546&gpsTopLongitude=2.662193&gpsBotLongitude=1.898879&gpsBotLatitude=48.572554&nbStations=0&bikes=yes"
	stationDetailsEndpoint = "/secured/searchStation?disponibility=yes"
)

func (api *VelibApiClient) GetVelibAtStations(stationName string) ([]stationDetailApiResponse, error) {
	body := getStationBody{
		StationName:   stationName,
		Disponibility: "yes",
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse body into JSON: %w", err)
	}
	headers := map[string]string{"Authorization": "Basic bW9iYTokMnkkMTAkRXNJYUk2LkRsZTh1elJGLlZGTEVKdTJ5MDJkc2xILnY3cUVvUkJHZ041MHNldUZpUkU1Ny4",
		"Content-Type": "application/json"}
	res, err := api.client.Post(baseUrl+stationDetailsEndpoint, headers, string(bodyJson))
	if err != nil {
		return nil, fmt.Errorf("cannot fetch Station details into JSON: %w", err)
	}
	if res.Status != 200 {
		return nil, fmt.Errorf("code unexpected response from Velib API %d when getting details about station: %s",
			res.Status, res.Body)
	}
	var respJson []stationDetailApiResponse
	err = json.Unmarshal([]byte(res.Body), &respJson)
	if err != nil {
		return nil, err
	}
	return respJson, nil
}

func (api *VelibApiClient) GetAllStations() ([]VelibApiEntity, error) {
	headers := map[string]string{"Authorization": "Basic bW9iYTokMnkkMTAkRXNJYUk2LkRsZTh1elJGLlZGTEVKdTJ5MDJkc2xILnY3cUVvUkJHZ041MHNldUZpUkU1Ny4"}
	res, err := api.client.Get(baseUrl+allStationsEndpoint, headers, "")
	if err != nil {
		return nil, err
	}
	var respJson []stationsApiResponse
	err = json.Unmarshal([]byte(res.Body), &respJson)
	if err != nil {
		return nil, err
	}
	var cleanedStation []VelibApiEntity
	for _, v := range respJson {
		if !strings.Contains(v.Station.Code, "_relais") {
			code, err := strconv.Atoi(v.Station.Code)
			if err != nil {
				return nil, fmt.Errorf("Cannot casr code to int: %w", err)
			}
			cleanedStation = append(cleanedStation, VelibApiEntity{
				Code:      code,
				Latitude:  v.Station.Gps.Latitude,
				Longitude: v.Station.Gps.Longitude,
				Name:      v.Station.Name,
			})
		}
	}
	return cleanedStation, nil
}
