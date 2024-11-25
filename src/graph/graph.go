package graph

import (
	"context"
	"fmt"
	"github.com/machinebox/graphql"
)

type PoolCreated struct {
	BlockNumber    string `json:"blockNumber"`
	Token0         string `json:"token0"`
	Token1         string `json:"token1"`
	Fee            int32  `json:"fee"`
	TickSpacing    int32  `json:"tickSpacing"`
	Pool           string `json:"pool"`
	BlockTimestamp string `json:"blockTimestamp"`
}

type poolCreatedsResponse struct {
	PoolCreateds []PoolCreated `json:"poolCreateds"`
}

type GraphClient struct {
	*graphql.Client
}

func (client GraphClient) QueryLastBlockNumber() string {
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
		fmt.Printf("查询失败最新区块号失败: %v\n", err)
	}

	return resp.PoolCreateds[0].BlockNumber
}

func (client GraphClient) QueryPoolCreatedsByPage(pageSize int, startBlockNumber string) []PoolCreated {
	// 初始化分页查询
	ctx := context.Background()
	poolCreateds := make([]PoolCreated, 0)
	skip := 0
	query := `
			query ($startBlockNumber: String!, $first: Int!, $skip: Int!) {
			poolCreateds(
				where: { blockNumber_gt: $startBlockNumber }
				first: $first
				skip: $skip
				orderBy: blockNumber
				orderDirection: asc
			) {
				blockNumber
				token0
				token1
				fee
				tickSpacing
				pool
				blockTimestamp
			}
		}
		`

	// 循环分页查询
	for {
		// 创建 GraphQL 请求
		req := graphql.NewRequest(query)
		req.Var("startBlockNumber", startBlockNumber)
		req.Var("first", pageSize)
		req.Var("skip", skip)

		// 执行查询
		var resp = poolCreatedsResponse{}
		if err := client.Run(ctx, req, &resp); err != nil {
			fmt.Printf("分页查询失败: %v\n", err)
		}

		// 无结果，退出循环，存在则将结果填充到 `poolCreateds`中
		if len(resp.PoolCreateds) == 0 {
			break
		}

		// 输出当前页数据
		// for _, poolCreated := range resp.PoolCreateds {
		// 	fmt.Printf("Block: %s, Token0: %s, Token1: %s, Fee: %d, TickSpacing: %d, Pool: %s\n",
		// 		poolCreated.BlockNumber, poolCreated.Token0, poolCreated.Token1, poolCreated.Fee, poolCreated.TickSpacing, poolCreated.Pool)
		// }

		// 更新下一个分页的起点
		poolCreateds = append(poolCreateds, resp.PoolCreateds...)
		skip += pageSize
	}

	return poolCreateds
}
