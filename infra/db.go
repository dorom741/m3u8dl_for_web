package infra

import (
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DataDB *gorm.DB

func InitGORM(dsn string, customerLogger logger.Writer) {
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

	var (
		driverName     = strings.Split(dsn, ":")[0]
		dataSourceName = dsn[len(driverName)+1:]
		err            error
		dialector      gorm.Dialector
	)

	switch driverName {
	case "mysql":
		dialector = mysql.Open(dataSourceName)
	case "sqlite":
		dialector = sqlite.Open(dataSourceName)

	}

	// 调用 Open 方法，传入驱动名和连接字符串
	DataDB, err = gorm.Open(dialector, &gorm.Config{
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
