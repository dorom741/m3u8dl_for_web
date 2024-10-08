package controller

import "m3u8dl_for_web/service"

var TaskControllerInstance = NewTaskController(service.M3u8dlServiceInstance, service.SubtitleWorkerServiceInstance)
