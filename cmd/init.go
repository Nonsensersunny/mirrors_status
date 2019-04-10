package main

import (
	"github.com/gin-gonic/gin"
	"mirrors_status/cmd/config"
	"mirrors_status/cmd/infrastructure"
	"mirrors_status/cmd/log"
)

type App struct {
	serverConfig *configs.ServerConf
}

func Init() (app App) {
	log.Info("Initializing APP")
	var sc configs.ServerConf
	serverConfig := sc.GetConfig()
	app = App{
		serverConfig: serverConfig,
	}

	InitInfluxDB(*serverConfig)
	return
}

func InitInfluxDB(config configs.ServerConf) {
	host := config.InfluxDB.Host
	port := config.InfluxDB.Port
	dbName := config.InfluxDB.DBName
	username := config.InfluxDB.Username
	password := config.InfluxDB.Password
	err := infrastructure.InitInfluxdbClient(host, port, dbName, username, password)
	if err != nil {
		log.Errorf("Err connecting influxdb:%v", config.InfluxDB)
	}
}

//func GetData(c *gin.Context) {
//	client := infrastructure.GetInfluxdbClient()
//	c.JSON(200, gin.H{
//		"res": client.Exec(""),
//	})
//}

func main() {
	app := Init()

	r := gin.Default()
	//r.GET("/", GetData)
	r.Run(":" + app.serverConfig.Http.Port)
}
