package clients

import (
	"encoding/json"
	"fmt"
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/RomainMichau/NordVpn_Server_Picker/srvpicker"
	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	nordVpnHttpsProxyPort  = 89
	baseUrl                = "https://www.velib-metropole.fr/api"
	allStationsEndpoint    = "/map/details?zoomLevel=1&gpsTopLatitude=49.05546&gpsTopLongitude=2.662193&gpsBotLongitude=1.898879&gpsBotLatitude=48.572554&nbStations=0&bikes=yes"
	stationDetailsEndpoint = "/secured/searchStation?disponibility=yes"
)

type VelibApiClient struct {
	channel          cloudscraper.CloudScrapper
	apiToken         string
	RespChan         chan cycletls.Response
	nordVpnSrvPicker *srvpicker.SrvPicker
	useProxy         bool
	proxyUsername    string
	proxyPassword    string
	proxyUrl         string
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
		useProxy: false,
	}
	return &api
}
func InitVelibApiUsingNordVpnProxy(apiToken string, proxyUsername string, proxyPassword string,
	proxyCountry string) *VelibApiClient {
	nordPickerOptions := srvpicker.Options{
		Country: proxyCountry,
		Feature: "proxy_ssl",
	}
	nordServerPicker := srvpicker.Init(&nordPickerOptions)
	channel, _ := cloudscraper.Init(false, true)
	api := VelibApiClient{
		channel:          *channel,
		apiToken:         apiToken,
		RespChan:         channel.RespChan(),
		nordVpnSrvPicker: nordServerPicker,
		useProxy:         true,
		proxyUsername:    proxyUsername,
		proxyPassword:    proxyPassword,
	}
	return &api
}

func (api *VelibApiClient) getProxyUrl() (string, error) {
	rand.Seed(time.Now().UnixNano())
	if !api.useProxy {
		return "", nil
	}
	if api.proxyUrl == "" {
		proxyServers, err := api.nordVpnSrvPicker.GetServers()
		if err != nil {
			return "", fmt.Errorf("failed to get a NordVpn proxy server: %w", err)
		}
		if len(proxyServers) == 0 {
			return "", fmt.Errorf("no NordVpn proxy server returned with param")
		}
		rnd := rand.Intn(len(proxyServers))
		proxyServerDomain := proxyServers[rnd].Domain
		fmt.Printf("Will use proxy %s\n", proxyServerDomain)
		api.proxyUrl = fmt.Sprintf("https://%s:%s@%s:%d", api.proxyUsername, api.proxyPassword, proxyServerDomain, nordVpnHttpsProxyPort)
	}
	return api.proxyUrl, nil
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
	proxyUrl, err := api.getProxyUrl()
	if err != nil {
		return err
	}
	options := cycletls.Options{
		Body:    string(bodyJson),
		Headers: headers,
		Timeout: 1000,
		Proxy:   proxyUrl,
	}
	api.channel.Queue(baseUrl+stationDetailsEndpoint, options, "POST")
	return nil
}

func (api *VelibApiClient) ParseGetStationDetailResponse(resp cycletls.Response) (StationDetailApiResponse, error) {
	if resp.Status != 200 {
		if strings.Contains(strings.ToLower(resp.Body), "proxy") {
			fmt.Println("Proxy error. Changing proxy address")
		}
		return StationDetailApiResponse{}, fmt.Errorf("code unexpected response from Velib API %d when getting details about station: %s",
			resp.Status, resp.Body)
	}
	var respJson []StationDetailApiResponse
	err := json.Unmarshal([]byte(resp.Body), &respJson)
	if err != nil {
		return StationDetailApiResponse{}, err
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

func (api *VelibApiClient) GetAllStations() ([]StationApiEntity, error) {
	headers := map[string]string{"Authorization": fmt.Sprintf("Basic %s", api.apiToken)}
	proxyUrl, err := api.getProxyUrl()
	if err != nil {
		return nil, err
	}
	options := cycletls.Options{
		Headers: headers,
		Proxy:   proxyUrl,
	}
	res, err := api.channel.Do(baseUrl+allStationsEndpoint, options, "GET")
	if err != nil {
		api.proxyUrl = ""
		return nil, err
	}
	if res.Status != 200 {
		if strings.Contains(strings.ToLower(res.Body), "proxy") {
			fmt.Println("Proxy error. Changing proxy address")
		}
		return nil, fmt.Errorf("unexpected response code %d. Body: %s", res.Status, res.Body)
	}
	var respJson []stationsApiResponse
	err = json.Unmarshal([]byte(res.Body), &respJson)
	if err != nil {
		return nil, fmt.Errorf("failed to jsonify %s: %w", res.Body, err)
	}
	var cleanedStation []StationApiEntity
	for _, v := range respJson {
		if !strings.Contains(v.Station.Code, "_relais") {
			code, err := strconv.Atoi(v.Station.Code)
			if err != nil {
				return nil, fmt.Errorf("Cannot cast code to int: %w", err)
			}
			cleanedStation = append(cleanedStation, StationApiEntity{
				Code:      code,
				Latitude:  v.Station.Gps.Latitude,
				Longitude: v.Station.Gps.Longitude,
				Name:      v.Station.Name,
			})
		}
	}
	if cleanedStation == nil {
		print("blk")
	}
	return cleanedStation, nil
}
