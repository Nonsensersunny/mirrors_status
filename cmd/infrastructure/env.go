package infrastructure

import (
	"mirrors_status/cmd/modules/db/influxdb"
)

var (
	influxdbClient *influxdb.Client
)

func GetInfluxdbClient() *influxdb.Client {
	return influxdbClient
}

func InitInfluxdbClient(host string, port int,
	dbname, username, password string) (err error) {
	influxdbClient = &influxdb.Client{
		Host:     host,
		Port:     port,
		DbName:   dbname,
		Username: username,
		Password: password,
	}
	return
}
