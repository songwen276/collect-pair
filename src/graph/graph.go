package graph

import (
	"collectv2v3uniswap/src/config"
	"context"
	"fmt"
	"github.com/machinebox/graphql"
)

type PoolCreated struct {
	BlockNumber string `json:"blockNumber"`
	Token0      string `json:"token0"`
	Token1      string `json:"token1"`
	Fee         int    `json:"fee"`
	TickSpacing int    `json:"tickSpacing"`
	Pool        string `json:"pool"`
}

type poolCreatedsResponse struct {
	PoolCreateds []PoolCreated `json:"poolCreateds"`
}

// 初始化 GraphQL 客户端
var client = graphql.NewClient(config.ConfigCache.Graph.Url)

func QueryLastBlockNumber() string {
	// 构建 GraphQL 查询
	ctx := context.Background()
	query := `
			query {
				poolCreateds(first: 1, orderBy: blockNumber, orderDirection: desc) {
                  blockNumber
                }
			}
		`
	req := graphql.NewRequest(query)

	// 执行查询
	var resp = poolCreatedsResponse{}
	if err := client.Run(ctx, req, &resp); err != nil {
		fmt.Printf("查询失败最新区块号失败: %v", err)
	}

	return resp.PoolCreateds[0].BlockNumber
}

func QueryPoolCreatedsByPage(pageSize int, startBlockNumber string) []PoolCreated {
	// 初始化分页查询
	ctx := context.Background()
	poolCreateds := make([]PoolCreated, 0)
	query := `
			query ($first: Int, $blockNumber: String) {
				poolCreateds(first: $first, where: { blockNumber_gt: $blockNumber }, orderBy: blockNumber, orderDirection: asc) {
					blockNumber
					token0
					token1
					fee
					tickSpacing
					pool
				}
			}
		`

	// 循环分页查询
	for {
		// 创建 GraphQL 请求
		req := graphql.NewRequest(query)
		req.Var("first", pageSize)
		req.Var("blockNumber", startBlockNumber)

		// 执行查询
		var resp = poolCreatedsResponse{}
		if err := client.Run(ctx, req, &resp); err != nil {
			fmt.Printf("分页查询失败: %v", err)
		}

		// 无结果，退出循环，存在则将结果填充到 `poolCreateds`中
		if len(resp.PoolCreateds) == 0 {
			break
		} else {
			poolCreateds = append(poolCreateds, resp.PoolCreateds...)
		}

		// 输出当前页数据
		// for _, poolCreated := range resp.PoolCreateds {
		// 	fmt.Printf("Block: %s, Token0: %s, Token1: %s, Fee: %d, TickSpacing: %d, Pool: %s\n",
		// 		poolCreated.BlockNumber, poolCreated.Token0, poolCreated.Token1, poolCreated.Fee, poolCreated.TickSpacing, poolCreated.Pool)
		// }

		// 更新下一个分页的起点为当前页最后一条记录的 `blockNumber`
		startBlockNumber = resp.PoolCreateds[len(resp.PoolCreateds)-1].BlockNumber
	}

	return poolCreateds
}
