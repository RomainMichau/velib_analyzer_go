package main

import (
	"flag"
	"fmt"
	"github.com/RomainMichau/velib_finder/clients"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type Params struct {
	DbHostname       string
	ApiToken         string
	DbPassword       string
	DbUsername       string
	DbPort           int
	DbName           string
	IntervalSec      int
	displayPubIp     bool
	Verbose          bool
	requestMaxFreqMs int
}

func getMyPublicIp() (string, error) {
	resp, err := http.Get("https://ifconfig.me/")
	if err != nil {
		return "", fmt.Errorf("Failed to query ifconfig.me: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read ifconfig.me response: %w", err)
	}
	return string(body), nil
}

func main() {
	params, err := parseParams()
	if params.Verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	if params.displayPubIp {
		ip, err := getMyPublicIp()
		if err != nil {
			panic(err)
		} else {
			log.Infof("Public IP: %s\n", ip)
		}
	}
	if err != nil {
		panic(err)
	}
	sql, _ := clients.InitSql(params.DbPassword, params.DbHostname, params.DbUsername, params.DbName, params.DbPort)
	api := clients.InitVelibApi(params.ApiToken)
	exporter := InitDbExporter(api, sql, 200, time.Duration(params.requestMaxFreqMs))
	for {
		log.Infof("Running DB export")
		err := exporter.RunExport()
		if err != nil {
			log.Errorf("Fail to run DB export: %s", err.Error())
		} else {
			log.Infof("DB export ran successfully")
		}
		time.Sleep(time.Duration(params.IntervalSec) * time.Second)
	}
}

func parseParams() (*Params, error) {
	velibApiToken := flag.String("velib_api_token", "", "Used to query velib API")
	dbHostname := flag.String("db_hostname", "", "DB Hostname")
	dbPassword := flag.String("db_password", "", "DB Password")
	dbName := flag.String("db_name", "", "DB Name")
	dbUserName := flag.String("db_user", "", "DB username")
	dbPort := flag.Int("db_port", 5432, "DB username")
	intervalSeconds := flag.Int("interval_sec", 600, "Run interval in seconds")
	verbose := flag.Bool("log", false, "verbose")
	requestMaxFreqMs := flag.Int("request_max_freq_ms", 50, "Minimum nb on ms between 2 request to "+
		"velib API ")
	displayPubIp := flag.Bool("show_ip", false, "Log level")
	flag.Parse()
	if *dbHostname == "" {
		return nil, fmt.Errorf("db_hostname param required")
	}
	if *velibApiToken == "" {
		return nil, fmt.Errorf("velib_api_token param required")
	}
	if *dbPort == 0 {
		return nil, fmt.Errorf("db_port param required")
	}
	if *intervalSeconds == 0 {
		return nil, fmt.Errorf("interval_sec param required")
	}
	if *dbPassword == "" {
		return nil, fmt.Errorf("db_password param required")
	}
	if *dbUserName == "" {
		return nil, fmt.Errorf("db_user param required")
	}
	if *dbName == "" {
		return nil, fmt.Errorf("db_name param required")
	}
	return &Params{
		DbHostname:       *dbHostname,
		ApiToken:         *velibApiToken,
		DbPassword:       *dbPassword,
		DbUsername:       *dbUserName,
		DbPort:           *dbPort,
		DbName:           *dbName,
		Verbose:          *verbose,
		IntervalSec:      *intervalSeconds,
		displayPubIp:     *displayPubIp,
		requestMaxFreqMs: *requestMaxFreqMs,
	}, nil
}
