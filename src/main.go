package main

import (
	"collectv2v3uniswap/src/config"
	"collectv2v3uniswap/src/graph"
	"collectv2v3uniswap/src/pair"
	"fmt"
	"strconv"
	"time"
)

func main() {
	// 创建一个 1 小时的定时器
	duration := time.Duration(config.ConfigCache.Local.TickerTime) * time.Second
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C: // 每当时间到达 1 小时触发
			// 查询最新区块号，判断最新区块号是否与本地区块一致，一致则跳过
			lastBlockNumber := graph.QueryLastBlockNumber()
			if lastBlockNumber == config.ConfigCache.Local.Number {
				fmt.Printf("lastBlockNumber = localBlockNumber = %s , 无新增区块数据事件，跳过处理\n", lastBlockNumber)
				continue
			} else {
				fmt.Printf("存在新增区块数据事件，开始处理，区块范围: (%s - %s]\n", config.ConfigCache.Local.Number, lastBlockNumber)
			}

			// 查询新增的区块数据事件数据
			startBlockNumber, _ := strconv.ParseUint(config.ConfigCache.Local.Number, 10, 64)
			startBlockNumber++
			poolCreateds := graph.QueryPoolCreatedsByPage(config.ConfigCache.Local.PageSize, strconv.FormatUint(startBlockNumber, 10))
			fmt.Printf("查询新增区块数据事件成功，lenth: %d, poolCreateds: %v\n", len(poolCreateds), poolCreateds)

			// 判断该区块存在事件数据，进行后续处理
			if len(poolCreateds) > 0 {
				// 查询数据库中已存在router的最大pair_index
				maxPairIndex, err := pair.GetMaxPairIndexByRouter(config.ConfigCache.Graph.Swap_v3_address)
				if err != nil {
					fmt.Printf("查询数据库失败，err: %v\n", err)
					break loop
				}

				// 循环处理新增的区块数据事件
				groupedMap := make(map[string][]pair.ArbitragePair)
				for _, poolCreated := range poolCreateds {
					// 构造pair对象
					maxPairIndex++
					arbitragePair := pair.ArbitragePair{
						Router:       config.ConfigCache.Graph.Swap_v3_address,
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

					// 以 BlockNumber 为键，将 PoolCreated 加入对应的切片中
					groupedMap[poolCreated.BlockNumber] = append(groupedMap[poolCreated.BlockNumber], arbitragePair)
				}

				fmt.Printf("分组处理成功，groupedMap: %v\n", groupedMap)

				// 分组插入数据库
				for blockNumber, pairs := range groupedMap {
					err = pair.InsertArbitragePairsBatch(pairs)
					if err != nil {
						fmt.Printf("插入当前区块uniswap数据到数据库失败，err: %v\n", err)
						break loop
					}

					// 新增的区块的所有事件数据插入数据库成功后，更新本地区块号
					config.ConfigCache.Local.Number = blockNumber
					config.SaveConfig(config.ConfigCache)
					fmt.Printf("插入当前区块uniswap数据到数据库成功，blockNumber: %s\n", blockNumber)
				}
			}

		}
	}
}
