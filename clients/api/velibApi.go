package api

import (
	"fmt"
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
	"github.com/RomainMichau/velib_analyzer_go/clients"
	jsoniter "github.com/json-iterator/go"
	"strconv"
	"strings"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type VelibApiClient struct {
	channel  cloudscraper.CloudScrapper
	apiToken string
	RespChan chan cycletls.Response
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

type StationDetailApiResponse struct {
	stationsApiResponse
	Bikes []struct {
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

func InitVelibApi(apiToken string) *VelibApiClient {
	channel, _ := cloudscraper.Init(false, true)
	api := VelibApiClient{
		channel:  *channel,
		apiToken: apiToken,
		RespChan: channel.RespChan(),
	}
	return &api
}

var (
	baseUrl                = "https://www.velib-metropole.fr/api"
	allStationsEndpoint    = "/map/details?zoomLevel=1&gpsTopLatitude=49.05546&gpsTopLongitude=2.662193&gpsBotLongitude=1.898879&gpsBotLatitude=48.572554&nbStations=0&bikes=yes"
	stationDetailsEndpoint = "/secured/searchStation?disponibility=yes"
)

func (api *VelibApiClient) GetVelibAtStations(stationName string) (StationDetailApiResponse, error) {
	body := getStationBody{
		StationName:   stationName,
		Disponibility: "yes",
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return StationDetailApiResponse{}, fmt.Errorf("Cannot parse body into JSON: %w", err)
	}
	headers := map[string]string{"Authorization": fmt.Sprintf("Basic %s", api.apiToken),
		"Content-Type": "application/json"}
	res, err := api.channel.Post(baseUrl+stationDetailsEndpoint, headers, string(bodyJson))
	if err != nil {
		return StationDetailApiResponse{}, fmt.Errorf("failed to send request to velib api. %w", err)
	}
	return api.ParseGetStationDetailResponse(res)
}

func (api *VelibApiClient) QueueGetVelibRequest(stationName string) error {
	body := getStationBody{
		StationName:   stationName,
		Disponibility: "yes",
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("Cannot parse body into JSON: %w", err)
	}
	headers := map[string]string{"Authorization": fmt.Sprintf("Basic %s", api.apiToken),
		"Content-Type": "application/json"}
	options := cycletls.Options{
		Body:    string(bodyJson),
		Headers: headers,
		Timeout: 100,
	}
	api.channel.Queue(baseUrl+stationDetailsEndpoint, options, "POST")
	return nil
}

func (api *VelibApiClient) ParseGetStationDetailResponse(resp cycletls.Response) (StationDetailApiResponse, error) {
	if resp.Status != 200 {
		return StationDetailApiResponse{}, fmt.Errorf("code unexpected response from Velib API %d when getting details about station: %s",
			resp.Status, resp.Body)
	}
	var respJson []StationDetailApiResponse
	err := json.Unmarshal([]byte(resp.Body), &respJson)
	if err != nil {
		return StationDetailApiResponse{}, fmt.Errorf("failed to umarshal response: %s", resp.Body)

	}
	shorterStationNameSize := 10000
	shorterStationID := 0
	for i, station := range respJson {
		sz := len(station.Station.Name)
		if sz < shorterStationNameSize {
			shorterStationID = i
			shorterStationNameSize = sz
		}
	}
	return respJson[shorterStationID], nil
}

func (api *VelibApiClient) GetAllStations() ([]clients.VelibApiEntity, error) {
	headers := map[string]string{"Authorization": "Basic bW9iYTokMnkkMTAkRXNJYUk2LkRsZTh1elJGLlZGTEVKdTJ5MDJkc2xILnY3cUVvUkJHZ041MHNldUZpUkU1Ny4"}
	res, err := api.channel.Get(baseUrl+allStationsEndpoint, headers, "")
	if err != nil {
		return nil, err
	}
	var respJson []stationsApiResponse
	err = json.Unmarshal([]byte(res.Body), &respJson)
	if err != nil {
		return nil, fmt.Errorf("failed to jsonify %s: %w", res.Body, err)
	}
	var cleanedStation []clients.VelibApiEntity
	for _, v := range respJson {
		stationCode := strings.ReplaceAll(v.Station.Code, "_relais", "")
		code, err := strconv.Atoi(stationCode)
		if err != nil {
			return nil, fmt.Errorf("Cannot cast code to int: %w", err)
		}
		cleanedStation = append(cleanedStation, clients.VelibApiEntity{
			Code:      code,
			Latitude:  v.Station.Gps.Latitude,
			Longitude: v.Station.Gps.Longitude,
			Name:      v.Station.Name,
		})
	}
	if cleanedStation == nil {
		return nil, fmt.Errorf("velib API returned an empty list of station. Body: %s", res.Body)
	}
	return cleanedStation, nil
}
