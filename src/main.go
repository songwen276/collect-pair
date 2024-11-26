package main

import (
	"collect-pair/src/task"
)

func main() {

	// 启动定时任务，定时刷新动态配置
	go task.TimerGetDynamicConfig()

	// 循环监听配置变更通知，根据任务配置启动相应的任务
	for {
		select {
		case <-task.DynamicCfgNotice:
			// 获取任务信息
			for _, collectTask := range task.TaskMap {
				// 启动任务
				if collectTask.On && collectTask.Running == false {
					collectTask.Running = true
					go task.StartCollectTask(collectTask)
				}
			}
		}
	}

}
