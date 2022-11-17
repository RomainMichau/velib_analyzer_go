package main

import (
	"flag"
	"fmt"
	"github.com/RomainMichau/velib_finder/clients"
	"time"
)

type Params struct {
	DbHostname  string
	ApiToken    string
	DbPassword  string
	DbUsername  string
	DbPort      int
	DbName      string
	LogLevel    string
	IntervalSec int
}

func main() {
	params, err := parseParams()
	if err != nil {
		panic(err)
	}
	sql, _ := clients.InitSql(params.DbPassword, params.DbHostname, params.DbUsername, params.DbName, params.DbPort)
	api := clients.InitVelibApi(params.ApiToken)
	exporter := InitDbExporter(api, sql, 200, 10)
	for {
		fmt.Println("Running DB export")
		err := exporter.RunExport()
		if err != nil {
			fmt.Println("Fail to run DB export: %w", err)
		} else {
			fmt.Println("DB export ran successfully")
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
	logLevel := flag.String("log", "INFO", "Log level")
	flag.Parse()
	if *logLevel == "" {
		return nil, fmt.Errorf("log param required")
	}
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
		DbHostname:  *dbHostname,
		ApiToken:    *velibApiToken,
		DbPassword:  *dbPassword,
		DbUsername:  *dbUserName,
		DbPort:      *dbPort,
		DbName:      *dbName,
		LogLevel:    *logLevel,
		IntervalSec: *intervalSeconds,
	}, nil
}
