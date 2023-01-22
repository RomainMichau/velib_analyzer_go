package main

import (
	"flag"
	"fmt"
	"github.com/RomainMichau/velib_analyzer_go/clients/api"
	"github.com/RomainMichau/velib_analyzer_go/clients/database"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type Params struct {
	DbHostname     string
	ApiToken       string
	DbPassword     string
	DbUsername     string
	DbPort         int
	DbName         string
	IntervalSec    int
	displayPubIp   bool
	Verbose        bool
	requestMaxFreq int
	noRunSync      bool
	apiPort        int
}

func (p *Params) print() {
	log.Infof("====================== PARAM ======================")
	log.Infof("DB host %s:%d", p.DbHostname, p.DbPort)
	log.Infof("DB_name: %s", p.DbName)
	log.Infof("DB_username: %s", p.DbUsername)
	log.Infof("Waiting time btw 2 runs: %d sec", p.IntervalSec)
	log.Infof("HTTP request max freq: %d", p.requestMaxFreq)
	log.Infof("Run sync disabled: %t", p.noRunSync)
	log.Infof("====================================================")
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
	if err != nil {
		panic(fmt.Errorf("failed to parse param: %w", err))
	}
	params.print()
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
			log.Infof("Public IP: %s", ip)
		}
	}
	sql, _ := database.InitDatabase(params.DbPassword, params.DbHostname, params.DbUsername, params.DbName, params.DbPort)
	err = sql.PostSync()
	if err != nil {
		return
	}
	metric := Metrics{failureCount: 0}
	controller := InitController(sql, &metric)
	go controller.Run(params.apiPort)
	velibApi := api.InitVelibApi(params.ApiToken)
	exporter := InitDbExporter(velibApi, sql, 200, time.Duration(1000/params.requestMaxFreq),
		time.Second*10)
	for {
		if !params.noRunSync {
			log.Infof("Running DB export")
			err := exporter.RunExport()
			if err != nil {
				log.Errorf("Fail to run DB export: %s", err.Error())
				metric.reportFailure()
			} else {
				log.Infof("DB export ran successfully")
			}
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
	verbose := flag.Bool("verbose", false, "verbose")
	apiPort := flag.Int("api_port", 80, "verbose")
	noRunSync := flag.Bool("no_run_sync", false, "run sync")
	requestMaxFreqMs := flag.Int("request_max_freq", 10, "Max request to API per second"+
		"velib API ")
	displayPubIp := flag.Bool("show_ip", false, "Log level")
	flag.Parse()
	log.Infof(*dbHostname)
	if *dbHostname == "" {
		return nil, fmt.Errorf("db_hostname param required")
	}
	if *velibApiToken == "" {
		return nil, fmt.Errorf("velib_api_token param required")
	}
	if *dbPort == 0 {
		return nil, fmt.Errorf("db_port param required")
	}
	if *intervalSeconds <= 0 {
		return nil, fmt.Errorf("interval_sec param required and cannot be <= 0")
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
	if *requestMaxFreqMs > 50 {
		return nil, fmt.Errorf("request_max_freq cannot be > 50")
	}
	return &Params{
		DbHostname:     *dbHostname,
		ApiToken:       *velibApiToken,
		DbPassword:     *dbPassword,
		DbUsername:     *dbUserName,
		DbPort:         *dbPort,
		DbName:         *dbName,
		Verbose:        *verbose,
		IntervalSec:    *intervalSeconds,
		displayPubIp:   *displayPubIp,
		requestMaxFreq: *requestMaxFreqMs,
		noRunSync:      *noRunSync,
		apiPort:        *apiPort,
	}, nil
}
