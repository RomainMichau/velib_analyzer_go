# Velib analyzer go
## Feature
### Exporter
Using the velib API this service will push velib info into an SQL DB
Including:
- List of stations
- List of Velib
- Which velib was docked at which station at which time
### API
Offer an HTTP Api to query aggregated data from the database
## Requirements
### Velib API token
Get a Velib API token.  
Can be obtained by reading HTTPS calls made by the phone app.  
You can do it using an emulator and [mitmproxy](https://borntocode.fr/mitmproxy-analyser-le-trafic-de-vos-applications-mobiles/). 
The mobile app don't use cert pinning, so it's easy
### SQL Database
Create a postgres DB using this [sql script](sql/db.sql). Require postgres > 9.4
## Run locally
### Build and run
! You should use a VPN when doing this as velib API is protected by cloudflare. You don't want cloudflare banning your IP

| Name                               | Mandatory | Default Value | Description                                                                                                | Example value     |
|------------------------------------|-----------|---------------|------------------------------------------------------------------------------------------------------------|-------------------|
| --velib_api_token \<token>         | true      |               | Token used to query the velib API without the 'Basic' prefix                                               |                   |
| --db_hostname \<hostname>          | true      |               | Hostname of your database                                                                                  | mydb.mydomain.com |
| --db_port \<port>                  | false     | 5432          | Port of your database                                                                                      | 5432              |
| --db_name \<db_name>               | true      |               | Name of your database                                                                                      | velib_analyzer    |
| --db_user \<username>              | true      |               | Username of the database. Must have rw access                                                              | postgres          |
| --db_password \<password>          | true      |               | Password of your DB                                                                                        |                   |
| --interval_sec \<interval>         | false     | 600           | Interval in second between 2 synchronization of the DB and the API                                         | 600               |
| --verbose                          | false     |               | If present will display more logs                                                                          |                   |
| --apiPort <port>                   | false     | 80            | Port used by the velib_analyzer API                                                                        | 8080              |
| --no_run_sync                      | false     |               | If present the app will not run the velib API => DB sync. Will only serve the API                          |                   |
| --request_max_freq \<interval_sec> | false     | 10            | Max nb of request to velib API per seconds. Don't use a too big  value to not harm the API. Cannot be > 50 | 15                |
| --displayPubIp                     | false     |               | If present will display your pub IP when starting the service. Useful when using it with a VPN             |                   |

```shell
git clone git@github.com:RomainMichau/velib_analyzer_go.git
cd velib_analyzer_go
go build
./velib_analyzer_go --velib_api_token XXXXX --db_hostname mydb.mydomain.com --db_user postgres --db_name velib_analyzer --db_password P@$$w0rd --show_ip --verbose --api_port 8081
```