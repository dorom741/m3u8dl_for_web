package infra

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var DataDB *gorm.DB

func InitGORM(dbUrl string,customerLogger logger.Writer ) {
	var err error
	slowLogger := logger.New(
		//设置Logger
		customerLogger,
		logger.Config{
			//慢SQL阈值
			SlowThreshold: time.Millisecond,
			//设置日志级别，只有Warn以上才会打印sql
			LogLevel: logger.Error,
		},
	)
	// 调用 Open 方法，传入驱动名和连接字符串
	DataDB, err = gorm.Open(sqlite.Open(dbUrl), &gorm.Config{
		Logger: slowLogger,
	})
	// 检查是否有错误
	if err != nil {
		customerLogger.Printf("database connect fail:", err)
		return
	}
	// 打印成功信息
	customerLogger.Printf("database success!")
}
