package model

import (
	"github.com/orestonce/m3u8d"
	"gorm.io/gorm"

	"m3u8dl_for_web/infra"
)

const (
	StateAll   = 0
	StateReady = 1
	StateEnd   = 2
	StateError = 3
)

type TaskRecord struct {
	gorm.Model
	ID      uint   `gorm:"primaryKey;autoIncrement"` // 主键，自增
	Name    string `gorm:"not null"`                 // 名称，非空
	URL     string `gorm:"not null"`                 // URL，非空
	SaveDir string `gorm:"not null"`                 // 保存目录，非空
	State   int    `gorm:"not null"`                 // 状态，非空
	Info    string `gorm:"not null"`                 // 信息，非空
	Reult   string

	Headers map[string][]string `gorm:"-"`
}

func (taskRecord TaskRecord) TableName() string {
	return "task_record"
}

func (taskRecord TaskRecord) ToStartDownloadReq() m3u8d.StartDownload_Req {
	return m3u8d.StartDownload_Req{
		M3u8Url:                  taskRecord.URL,
		Insecure:                 true,
		SaveDir:                  taskRecord.SaveDir,
		FileName:                 taskRecord.Name,
		SkipTsExpr:               "",
		SetProxy:                 "",
		HeaderMap:                taskRecord.Headers,
		SkipRemoveTs:             false,
		ProgressBarShow:          false,
		ThreadCount:              4,
		SkipCacheCheck:           false,
		SkipMergeTs:              false,
		Skip_EXT_X_DISCONTINUITY: false,
		DebugLog:                 false,
	}
}

func (taskRecord *TaskRecord) Save() error {
	var (
		// count int64
		db = infra.DataDB.Model(taskRecord)
		// url = taskInfo.URL
	)

	// infra.DataDB.Model(&TaskRecord{}).Where("url = ?", url).Count(&count)

	result := db.Save(&taskRecord)
	if result.Error != nil {
		return result.Error
	}

	return nil

}

func (taskRecord *TaskRecord) Finish(result string) error {
	taskRecord.State = StateEnd

	if len(result) > 0 {
		taskRecord.State = StateError
		taskRecord.Reult = result

	}
	db := infra.DataDB.Model(taskRecord)
	return db.Save(taskRecord).Error
}

func (taskRecord *TaskRecord) GetNotWorkingWork() ([]TaskRecord, error) {

	taskRecordList := []TaskRecord{}
	db := infra.DataDB
	db.Error = nil
	db.Model(taskRecord)
	err := db.Where("state = ?", StateReady).Order("id").Find(&taskRecordList).Error
	if err != nil {
		return nil, err
	}

	return taskRecordList, nil
}
