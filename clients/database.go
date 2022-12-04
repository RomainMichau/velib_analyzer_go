package clients

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/url"
	"time"
)

type VelibSqlClient struct {
	connPool *pgxpool.Pool
}

var (
	SelectAllVelibCode = "SELECT velib_code FROM public.velibs"
	SelectVelibByCode  = `SELECT id, velib_code, electric
		FROM public.velibs WHERE velib_code = $1;`
	SelectLastStationForAllVelib = `select velib_code , station_code
		from (
		select
			velib_code, station_code,
			row_number() over (partition by velib_code order by timestamp desc) as index
		from velib_docked
		) as sr
		where index = 1`
	SelectAllStations = `SELECT id, station_code, station_name, long, lat, station_code
		FROM public.stations`
	SelectStationByCode = `SELECT station_name, long, lat
		FROM public.stations WHERE station_code = $1`
	SelectLastDockedStationForVelib = `SELECT "timestamp", station_code, available
		FROM public.velib_docked code where velib_code = $1 order by "timestamp" desc limit 1`
	SelectAllStationForVelib = `select "timestamp" , s.station_name, s.long, s.lat  from velib_docked as v join stations s on s.station_code = v.station_code  
		where velib_code = $1 order by "timestamp" desc`
	InsertVelib = `INSERT INTO public.velibs
		(velib_code, electric, run)
		VALUES($1, $2, $3);`
	InsertRun = `INSERT INTO public.run
		(run_time)
		VALUES($1) RETURNING id`
	InsertStation = `INSERT INTO public.stations
		(station_name, long, lat, station_code, run)
		VALUES($1, $2, $3, $4, $5)`
	InsertVelibDocked = `INSERT INTO public.velib_docked
		(velib_code, timestamp, station_code, run ,available)
		VALUES($1, $2, $3, $4, $5)`
	RegisterRunSuccess = `update run  set success  = true  , minor_issues_count = $1 where id = $2`
)

func InitSql(dbPassword, dbHostname, dbUsername, dbName string, dbPort int) (*VelibSqlClient, error) {
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		url.QueryEscape(dbUsername), url.QueryEscape(dbPassword), dbHostname, dbPort, dbName)
	dbpool, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("Unable to create connection pool: %v\n", err)
	}
	return &VelibSqlClient{
		connPool: dbpool,
	}, nil
}

func (sql *VelibSqlClient) GetAllVelibCode() []int {
	res := []int{}
	rows, _ := sql.connPool.Query(context.Background(), SelectAllVelibCode)
	for rows.Next() {
		var code int
		rows.Scan(&code)
		res = append(res, code)
	}
	return res
}

func (sql *VelibSqlClient) GetLastStationForAllVelib() (map[int]int, error) {
	res := map[int]int{}
	rows, err := sql.connPool.Query(context.Background(), SelectLastStationForAllVelib)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch station for all velib: %w", err)
	}
	for rows.Next() {
		var velibCode int
		var stationCode int
		err := rows.Scan(&velibCode, &stationCode)
		if err != nil {
			return nil, err
		}
		res[velibCode] = stationCode
	}
	return res, nil
}

func (sql *VelibSqlClient) GetVelibByCode(code int) (*VelibSqlEntity, error) {
	row := sql.connPool.QueryRow(context.Background(), SelectVelibByCode, code)
	var id int
	var rcode int
	var electric bool
	err := row.Scan(&id, &rcode, &electric)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &VelibSqlEntity{
		Id:       id,
		Code:     rcode,
		Electric: electric,
	}, nil
}

func (sql *VelibSqlClient) GetLastDockedForVelib(velibCode int) (*VelibDockedSql, error) {
	time.Now()
	row := sql.connPool.QueryRow(context.Background(), SelectLastDockedStationForVelib, velibCode)
	var stationCode int
	var timestamp time.Time
	var available bool
	err := row.Scan(&timestamp, &stationCode, &available)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &VelibDockedSql{
		VelibCode:   velibCode,
		StationCode: stationCode,
		Available:   available,
		timeStamp:   timestamp,
	}, nil
}

func (sql *VelibSqlClient) GetStationByCode(code int) (*StationSqlEntity, error) {
	row := sql.connPool.QueryRow(context.Background(), SelectStationByCode, code)
	var name string
	var longitude float32
	var latitude float32
	err := row.Scan(&name, &longitude, &latitude)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &StationSqlEntity{
		Name:      name,
		Longitude: longitude,
		Latitude:  latitude,
		Code:      code,
	}, nil
}

func (sql *VelibSqlClient) GetAllStationsCode() (map[int]bool, error) {
	res := map[int]bool{}
	rows, err := sql.connPool.Query(context.Background(), SelectAllStations)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch all stations: %w", err)
	}
	for rows.Next() {
		var stationCode int
		err := rows.Scan(&stationCode, nil, nil, nil, nil, nil)
		if err != nil {
			return nil, err
		}
		res[stationCode] = true
	}
	return res, nil
}

func (sql *VelibSqlClient) InsertVelib(velibCode, run int, electric bool) error {
	ctx := context.Background()
	tx, err := sql.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, InsertVelib, velibCode, electric, run)
	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (sql *VelibSqlClient) InsertStation(name string, longitude, latitude float32, code, runId int) error {
	ctx := context.Background()
	tx, err := sql.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, InsertStation, name, longitude, latitude, code, runId)
	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (sql *VelibSqlClient) InsertVelibDocked(velibCode, stationCode, runId int, time time.Time, available bool) error {
	ctx := context.Background()
	tx, err := sql.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, InsertVelibDocked, velibCode, time, stationCode, runId, available)
	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (sql *VelibSqlClient) InsertRun(time time.Time) (int, error) {
	ctx := context.Background()
	tx, err := sql.connPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(ctx, InsertRun, time)
	var id int
	err = row.Scan(&id)
	if err != nil {
		return 0, err
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (sql *VelibSqlClient) RegisterSuccess(runId, minorIssueCount int) error {
	ctx := context.Background()
	tx, err := sql.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, RegisterRunSuccess, minorIssueCount, runId)
	if err != nil {
		return fmt.Errorf("failed to register Success in SQL: %w", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("failed to commit register Success in SQL: %w", err)
	}
	return nil
}
