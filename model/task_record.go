package model

import (
	"gorm.io/gorm"
	"m3u8dl_for_web/infra"
)

const (
	StateAll   = 0
	StateReady = 1
	StateEnd   = 2
	StateError = 3
)

type TaskRecord[I any, O any] struct {
	gorm.Model
	ID     uint `gorm:"primaryKey;autoIncrement"` // 主键，自增
	Input  I    `gorm:"serializer:json"`
	Output O    `gorm:"serializer:json"`

	State  int `gorm:"not null"` // 状态，非空
	Result string
	Type   string
}

func (taskRecord *TaskRecord[I, O]) TableName() string {
	return "task_record"
}

func (taskRecord *TaskRecord[I, O]) Save() error {

	// count int64
	db := infra.DataDB.Model(taskRecord)
	// url = taskInfo.URL

	// infra.DataDB.Model(&TaskRecord{}).Where("url = ?", url).Count(&count)

	result := db.Save(&taskRecord)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (taskRecord *TaskRecord[I, O]) Finish(result string) error {
	taskRecord.State = StateEnd

	if len(result) > 0 {
		taskRecord.State = StateError
		taskRecord.Result = result

	}
	db := infra.DataDB.Model(taskRecord)
	return db.Save(taskRecord).Error
}

func (taskRecord *TaskRecord[I, O]) GetNotWorkingWork() ([]TaskRecord[I, O], error) {
	var taskRecordList []TaskRecord[I, O]
	db := infra.DataDB
	db.Error = nil
	db.Model(taskRecord)
	err := db.Where("state = ?", StateReady).Order("id").Find(&taskRecordList).Error
	if err != nil {
		return nil, err
	}

	return taskRecordList, nil
}
