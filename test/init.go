package test 

import (
	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/service"

	"github.com/sirupsen/logrus"
)

func init() {
	conf.InitConf("../config.yaml")
	infra.InitLogger(conf.ConfigInstance)

	infra.InitGORM(conf.ConfigInstance.Server.Dsn, logrus.StandardLogger())
	infra.MustInitCache(conf.ConfigInstance.GetAbsCachePath())
	service.InitService(conf.ConfigInstance)
	err := infra.DataDB.AutoMigrate(&model.TaskRecord[struct{}, struct{}]{})
	if err != nil {
		panic(err)
	}
}
