package main

import (
	"errors"
	"os"
	"path"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/controller"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/service"
)

func main() {
	configFilePath, err := searchPath("./config.yaml")
	if err != nil {
		panic("配置文件不存在")
	}

	conf.InitConf(configFilePath)
	infra.MustInitCache(conf.ConfigInstance.Server.CacheDir)
	infra.InitLogger(conf.ConfigInstance)
	infra.InitGORM(conf.ConfigInstance.Server.Dsn, logrus.StandardLogger())
	if err := infra.InitHttpClientWithProxy(conf.ConfigInstance.Server.HttpClientProxy); err != nil {
		panic(err)
	}

	err = infra.DataDB.AutoMigrate(&model.TaskRecord[struct{}, struct{}]{})
	if err != nil {
		panic(err)
	}

	service.InitService(conf.ConfigInstance)
	controller.InitController()

	run(conf.ConfigInstance.Server.StaticPath)
}

func searchPath(filePath string) (string, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return filePath, nil
	}
	if !errors.Is(os.ErrNotExist, err) {
		return "", err
	}

	filePathOnParent := path.Join("re/", filePath)
	_, err = os.Stat(filePathOnParent)

	return filePathOnParent, err
}

func run(staticPath string) {
	r := gin.Default()
	_ = r.SetTrustedProxies([]string{"*"})

	// r.Use(middleware.LoggerMiddleware())

	pprof.Register(r)

	apiGroup := r.Group("/api")
	apiGroup.POST("/addM3u8dlTask", controller.TaskControllerInstance.AddM3u8dlTask)
	apiGroup.POST("/addGenerateSubtitleTask", controller.TaskControllerInstance.AddGenerateSubtitleTask)
	// apiGroup.POST("/addM3u8dlTaskByAria2", taskController.AddTaskByAria2)

	apiGroup.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	// r.Static("/", "./static")
	r.NoRoute(func(c *gin.Context) {
		fullPath := c.FullPath()
		if len(fullPath) == 0 || fullPath == "/" {
			fullPath = "index.html"
		}
		// 返回静态文件
		c.File(path.Join(staticPath, fullPath)) // 请确保该文件存在
	})

	// log.Infof("open http://127.0.0.1:2045/")
	err := r.Run(conf.ConfigInstance.Server.Listen)
	if err != nil {
		panic(err)
	}
}
