package pair

import (
	"collect-pair/src/mysqldb"
	"database/sql"
	"fmt"
	"time"
)

// ArbitragePair 定义表结构
type ArbitragePair struct {
	ID           uint64    `db:"id"`
	Router       string    `db:"router"`
	PairIndex    uint64    `db:"pair_index"`
	PairAddress  string    `db:"pair_address"`
	Token0       string    `db:"token0"`
	Token1       string    `db:"token1"`
	Amount0      string    `db:"amount0"`
	Amount1      string    `db:"amount1"`
	HasFlashLoan uint8     `db:"has_flash_loan"`
	Closed       uint8     `db:"closed"`
	GmtCreate    time.Time `db:"gmt_create"`
	GmtModified  time.Time `db:"gmt_modified"`
}

func InsertArbitragePairsBatch(arbitragePairs []ArbitragePair) error {
	db := mysqldb.GetMysqlDB()

	// 开启事务
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// 定义插入 SQL
	insert := `
        INSERT INTO arbitrage_pair (
            router, pair_index, pair_address, token0, token1, amount0, amount1, has_flash_loan, closed, gmt_create, gmt_modified
        ) VALUES (
            :router, :pair_index, :pair_address, :token0, :token1, :amount0, :amount1, :has_flash_loan, :closed, :gmt_create, :gmt_modified
        )`

	// 遍历数据进行插入
	for _, pair := range arbitragePairs {
		_, err = tx.NamedExec(insert, pair)
		if err != nil {
			tx.Rollback() // 回滚事务
			return fmt.Errorf("failed to insert data: %v\n", err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		tx.Rollback() // 回滚事务
		return fmt.Errorf("failed to commit transaction: %v\n", err)
	}

	return nil
}

func GetMaxPairIndexByRouter(router string) (uint64, error) {
	db := mysqldb.GetMysqlDB()

	var maxPairIndex sql.NullInt64
	query := `SELECT MAX(pair_index) AS max_pair_index FROM arbitrage_pair WHERE router = ?`

	err := db.Get(&maxPairIndex, query, router)
	if err != nil {
		return 0, fmt.Errorf("failed to query max pair_index for router %s: %v\n", router, err)
	}

	// 如果查询结果为 NULL，返回 0
	if !maxPairIndex.Valid {
		return 0, nil
	}

	return uint64(maxPairIndex.Int64), nil
}

func CountPair(router string, pairAddress string, token0 string, token1 string) (uint64, error) {
	db := mysqldb.GetMysqlDB()

	var count uint64
	query := `SELECT COUNT(1) FROM arbitrage_pair WHERE router = ? AND pair_address = ? AND token0 = ? AND token1 = ?`

	err := db.Get(&count, query, router, pairAddress, token0, token1)
	if err != nil {
		return 0, fmt.Errorf("failed to query is pair exists %s, %s, %s, %s: %v\n", router, pairAddress, token0, token1, err)
	}

	return count, nil
}
