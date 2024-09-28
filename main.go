package main

import (
	"errors"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

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

	staticPath, err := searchPath("./static")
	if err != nil {
		panic("静态资源文件不存在")
	}

	conf.InitConf(configFilePath)
	infra.InitLogger(*conf.ConfigInstance)
	infra.InitGORM("./data.db", infra.Logger)
	infra.DataDB.AutoMigrate(&model.TaskRecord{})
	run(staticPath)

}

func searchPath(filePath string) (string, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return filePath, nil
	}
	if errors.Is(os.ErrNotExist, err) {
		filePathOnParent := path.Join("../", filePath)
		_, err = os.Stat(filePathOnParent)
		return filePathOnParent, err

	}

	return "", err
}

func run(staticPath string) {
	r := gin.Default()
	taskController := controller.NewTaskController()
	go service.M3u8dlWorkerInstance.WorkerRun()

	apiGroup := r.Group("/api")
	apiGroup.POST("/addTask", taskController.AddTask)
	apiGroup.POST("/addTaskByAria2", taskController.AddTaskByAria2)

	apiGroup.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	// r.Static("/", "./static")
	r.NoRoute(func(c *gin.Context) {
		fullPath := c.FullPath()
		if len(fullPath) == 0 || fullPath == "/" {
			fullPath = "index.html"
		}

		c.Next()
		// 返回静态文件
		c.File(path.Join(staticPath, fullPath)) // 请确保该文件存在
	})

	log.Infof("open http://127.0.0.1:2045/static/")
	r.Run(conf.ConfigInstance.Server.Listen)
}
