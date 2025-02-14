package task

import (
	"collect-pair/src/config"
	mlog "collect-pair/src/log"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type TaskList struct {
	CollectTaskList []*CollectTask `json:"collect_task_list"`
}

func TimerGetDynamicConfig() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fetchDynamicConfig()
		}
	}
}

var DynamicCfgNotice = make(chan struct{})

func fetchDynamicConfig() {
	// 发送GET请求，获取最新的配置信息
	resp, err := http.Get(config.ConfigCache.Local.ConfigItemUrl)
	if err != nil {
		mlog.Logger.Errorf("http请求配置url失败，err: %v", err)
		return
	}
	defer resp.Body.Close() // 确保函数结束时关闭响应体

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		mlog.Logger.Errorf("读取http请求响应配置数据失败，err: %v", err)
		return
	}

	// 解析 JSON 数据
	list := &TaskList{}
	err = json.Unmarshal(body, &list)
	if err != nil {
		mlog.Logger.Errorf("解析配置数据失败，err: %v", err)
		return
	}

	// 读取任务配置保存到全局
	for _, collectTask := range list.CollectTaskList {
		// 任务已存在，更新配置
		if oriTask, ok := TaskMap[collectTask.ID]; ok {
			// 如果子图地址发生变化，则标记GraphUrlChanged变更
			if oriTask.GraphUrl != collectTask.GraphUrl {
				oriTask.GraphUrlChanged = true
				oriTask.GraphUrl = collectTask.GraphUrl
			}
			oriTask.Name = collectTask.Name
			oriTask.ContractAddress = collectTask.ContractAddress
			oriTask.HasFlashLoan = collectTask.HasFlashLoan
			oriTask.On = collectTask.On
		} else {
			// 新增任务
			collectTask.GraphUrlChanged = true
			TaskMap[collectTask.ID] = collectTask
		}

		// 打印解析后的结果
		mlog.Logger.Infof("解析配置数据成功，dynamicConfig: %v", *collectTask)
	}

	// 通知各个任务重新加载配置
	DynamicCfgNotice <- struct{}{}

}
