package task

import (
	"collect-pair/src/graph"
	mlog "collect-pair/src/log"
	"collect-pair/src/pair"
	"github.com/machinebox/graphql"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"time"
)

var TaskMap = make(map[string]*CollectTask)

type CollectTask struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ContractAddress string `json:"contract_address"`
	HasFlashLoan    int    `json:"has_flash_loan"`
	GraphUrl        string `json:"graph_url"`
	GraphUrlChanged bool
	GraphClient     *graph.GraphClient
	On              bool `json:"on"`
	Running         bool
	TickerTime      int `json:"ticker_time"`
	PageSize        int `json:"page_size"`
}

func (task *CollectTask) InitGraphClient() {
	task.GraphClient = &graph.GraphClient{Client: graphql.NewClient(task.GraphUrl)}
}

func (task *CollectTask) Stop() {
	task.Running = false
}

func StartCollectTask(taskInfo *CollectTask) {
	// 创建一个 1 小时的定时器
	duration := time.Duration(taskInfo.TickerTime) * time.Second
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	defer taskInfo.Stop()

loop:
	for {

		// 判断任务开关关闭直接退出
		if !taskInfo.On {
			mlog.Logger.Infof("[%s]任务关闭，退出", taskInfo.Name)
			return
		}

		// 判断graphUrl是否发生变化，如果发生变化则更新graphClient
		if taskInfo.GraphUrlChanged {
			taskInfo.InitGraphClient()
			taskInfo.GraphUrlChanged = false
		}

		select {
		case <-ticker.C:
			// 根据任务名称加载本地已处理区块记录文件
			filePath := taskInfo.Name + ".yaml"
			records, err := loadTaskRecords(filePath)
			if err != nil {
				mlog.Logger.Errorf("[%s]加载本地已处理区块记录文件失败，err: %v", taskInfo.Name, err)
				continue
			}

			// 查询最新区块号，判断最新区块号是否与本地区块一致，一致则跳过
			lastBlockNumber := taskInfo.GraphClient.QueryLastBlockNumber()
			if lastBlockNumber == records.LocalBlockNumber {
				mlog.Logger.Infof("[%s]lastBlockNumber = localBlockNumber = %s , 无新增区块数据事件，跳过处理", taskInfo.Name, lastBlockNumber)
				continue
			} else if lastBlockNumber < records.LocalBlockNumber {
				mlog.Logger.Infof("[%s]lastBlockNumber < localBlockNumber = %s , 本地区块数据异常或者无记录，退出", taskInfo.Name, lastBlockNumber)
				continue
			} else {
				mlog.Logger.Infof("[%s]存在新增区块数据事件，开始处理，区块范围: (%s - %s]", taskInfo.Name, records.LocalBlockNumber, lastBlockNumber)
			}

			// 查询新增的区块数据事件数据
			startBlockNumber, _ := strconv.ParseUint(records.LocalBlockNumber, 10, 64)
			poolCreateds := taskInfo.GraphClient.QueryPoolCreatedsByPage(taskInfo.PageSize, strconv.FormatUint(startBlockNumber, 10))
			mlog.Logger.Infof("[%s]查询新增区块数据事件成功，lenth: %d, poolCreateds: %v", taskInfo.Name, len(poolCreateds), poolCreateds)

			// 判断该区块存在事件数据，进行后续处理
			if len(poolCreateds) > 0 {
				// 查询数据库中已存在router的最大pair_index
				maxPairIndex, err := pair.GetMaxPairIndexByRouter(taskInfo.ContractAddress)
				if err != nil {
					mlog.Logger.Errorf("[%s]查询数据库失败，err: %v", taskInfo.Name, err)
					break loop
				}

				// 循环处理新增的区块数据事件
				groupedMap := make(map[string][]pair.ArbitragePair)
				uniqueNumbers := make(map[string]bool)
				filteredBlockNumbers := make([]string, 0)
				for _, poolCreated := range poolCreateds {
					// 判断该pair是否已存在，存在则跳过
					count, countErr := pair.CountPair(taskInfo.ContractAddress, poolCreated.Pool, poolCreated.Token0, poolCreated.Token1)
					if countErr != nil {
						break loop
					}

					if count > 0 {
						continue
					}

					// 构造pair对象
					maxPairIndex++
					arbitragePair := pair.ArbitragePair{
						Router:       taskInfo.ContractAddress,
						PairIndex:    maxPairIndex,
						PairAddress:  poolCreated.Pool,
						Token0:       poolCreated.Token0,
						Token1:       poolCreated.Token1,
						Amount0:      "0",
						Amount1:      "0",
						HasFlashLoan: 1,
						Closed:       0,
						GmtCreate:    time.Now().UTC(),
						GmtModified:  time.Now().UTC(),
					}

					// 为了保证后续处理顺序，将 BlockNumber 去重存入切片
					if _, exist := uniqueNumbers[poolCreated.BlockNumber]; !exist {
						filteredBlockNumbers = append(filteredBlockNumbers, poolCreated.BlockNumber)
						uniqueNumbers[poolCreated.BlockNumber] = true
					}

					// 以 BlockNumber 为键，将 待处理的 arbitragePair 加入对应的map中
					groupedMap[poolCreated.BlockNumber] = append(groupedMap[poolCreated.BlockNumber], arbitragePair)
				}

				mlog.Logger.Infof("[%s]分组处理成功，filteredBlockNumbers: %v，groupedMap: %v", taskInfo.Name, len(filteredBlockNumbers), len(groupedMap))

				// 分组插入数据库
			loopIntert:
				for _, blockNumber := range filteredBlockNumbers {
					pairs := groupedMap[blockNumber]
					err = pair.InsertArbitragePairsBatch(pairs)
					if err != nil {
						mlog.Logger.Errorf("[%s]插入当前区块pair数据到数据库失败，err: %v", taskInfo.Name, err)
						break loopIntert
					}

					// 新增的区块的所有事件数据插入数据库成功后，更新本地区块号
					records.LocalBlockNumber = blockNumber
					saveTaskRecords(filePath, records)
					mlog.Logger.Infof("[%s]插入当前区块pair数据到数据库成功，blockNumber: %s", taskInfo.Name, blockNumber)
				}
			}
		}
	}
}

type TaskRecords struct {
	LocalBlockNumber string `yaml:"localBlockNumber"`
}

func loadTaskRecords(filePath string) (*TaskRecords, error) {
	// 打开文件
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 创建 任务记录 实例
	var records TaskRecords

	// 使用 YAML 解码
	decoder := yaml.NewDecoder(f)
	if err = decoder.Decode(&records); err != nil {
		return nil, err
	}

	return &records, nil
}

func saveTaskRecords(filePath string, TaskRecords *TaskRecords) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	defer encoder.Close()

	return encoder.Encode(TaskRecords)
}
