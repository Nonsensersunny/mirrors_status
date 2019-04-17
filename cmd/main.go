package main

import (
	"github.com/gin-gonic/gin"
	"mirrors_status/cmd/infrastructure"
	"mirrors_status/pkg/business/cdn-checker"
	"mirrors_status/pkg/config"
	"mirrors_status/pkg/log"
	"mirrors_status/pkg/modules/db/influxdb"
	"mirrors_status/pkg/modules/db/mysql"
	"mirrors_status/pkg/modules/model"
	"mirrors_status/pkg/modules/service"
	"strconv"
)

type App struct {
	serverConfig *configs.ServerConf
	influxClient *influxdb.Client
	mysqlClient *mysql.Client
	cdnChecker *cdn_checker.CDNChecker
}

func Init() (app App) {
	log.Info("Initializing APP")
	var sc configs.ServerConf
	serverConfig := sc.GetConfig()
	app = App{
		serverConfig: serverConfig,
	}

	infrastructure.InitDB(*serverConfig)
	infrastructure.InitCDNCkecker(*app.serverConfig.CdnCkecker)
	app.influxClient = infrastructure.GetInfluxdbClient()
	app.mysqlClient = infrastructure.GetMySQLClient()
	app.cdnChecker = infrastructure.GetCdnChecker()

	infrastructure.InitScheme()
	return
}

func(app App) GetAllMirrors(c *gin.Context) {
	data := service.GetAllMirrors(app.influxClient)
	c.JSON(200, gin.H{
		"res": data,
	})
}

func(app App) GetAllMirrorsCdn(c *gin.Context) {
	data := service.GetAllMirrorsCdn(app.influxClient)
	c.JSON(200, gin.H{
		"res": data,
	})
}

func (app App) AddMirror(c *gin.Context) {
	var reqMirror model.MirrorsPoint
	err := c.ShouldBindJSON(&reqMirror)
	if err != nil {
		log.Errorf("Bind json found error:%v", err)
	}
	err = service.AddMirror(app.mysqlClient, app.influxClient, reqMirror)
	if err != nil {
		log.Errorf("Insert data found error:%v", err)
	}
	c.JSON(200, gin.H{
		"res": err,
	})
}

func (app App) AddMirrorCdn(c *gin.Context) {
	var reqMirrorCdn model.MirrorsCdnPoint
	err := c.ShouldBindJSON(&reqMirrorCdn)
	if err != nil {
		log.Errorf("Bind json found error:%v", err)
	}
	err = service.AddMirrorCdn(app.mysqlClient, app.influxClient, reqMirrorCdn)
	if err != nil {
		log.Errorf("Insert data found error:%v", err)
	}
	c.JSON(200, gin.H{
		"res": err,
	})
}

func (app App) TestApi(c *gin.Context) {
	query := c.PostForm("query")
	data := service.TestApi(app.influxClient, query)
	c.JSON(200, gin.H{
		"res": data,
	})
}

func (app App) SyncAllMirrors(c *gin.Context) {
	err := app.cdnChecker.CheckAllMirrors(app.mysqlClient, app.serverConfig.CdnCkecker)
	if err != nil {
		log.Errorf("Sync all mirror found error:%v", err)
	}
	c.JSON(200, gin.H{
		"res": err.Error(),
	})
}

func (app App) SyncMirror(c *gin.Context) {
	var reqMirror model.MirrorsPoint
	err := c.ShouldBindJSON(&reqMirror)
	res := reqMirror.Name
	if err != nil {
		log.Errorf("Bind json found error:%v", err)
		res = err.Error()
	}
	err = app.cdnChecker.CheckMirror(app.mysqlClient, reqMirror, app.serverConfig.CdnCkecker)
	if err != nil {
		log.Errorf("Sync mirror found error:%v", err)
		res = err.Error()
	}
	c.JSON(200, gin.H{
		"res": res,
	})
}

func main() {
	app := Init()
	r := gin.Default()

	r.GET("/mirrors", app.GetAllMirrors)
	r.GET("/mirrors_cdn", app.GetAllMirrorsCdn)

	r.POST("/mirrors", app.AddMirror)
	r.POST("/mirrors_cdn", app.AddMirrorCdn)

	r.POST("/test", app.TestApi)

	r.GET("/check", app.SyncAllMirrors)

	r.POST("/check", app.SyncMirror)



	r.Run(":" + strconv.Itoa(app.serverConfig.Http.Port))
}
