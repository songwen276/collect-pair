package mysqldb

import (
	"collectv2v3uniswap/src/config"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"sync"
	"time"
)

// DB 是全局数据库连接池
var (
	mysqldb *sqlx.DB
	once    sync.Once
)

// init 初始化数据库连接
func init() {
	once.Do(func() {
		// 构建 DSN (Data Source Name)
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", config.ConfigCache.Mysql.User, config.ConfigCache.Mysql.Password, config.ConfigCache.Mysql.Host, config.ConfigCache.Mysql.DBName)

		// 打开数据库连接
		var err error
		mysqldb, err = sqlx.Open("mysql", dsn)
		if err != nil {
			fmt.Printf("Error opening database", err)
		}

		// 配置连接池
		mysqldb.SetMaxOpenConns(config.ConfigCache.Mysql.MaxOpenConns)
		mysqldb.SetMaxIdleConns(config.ConfigCache.Mysql.MaxIdleConns)
		mysqldb.SetConnMaxLifetime(time.Duration(config.ConfigCache.Mysql.ConnMaxLifetime))

		// 验证连接
		if err = mysqldb.Ping(); err != nil {
			fmt.Printf("Error pinging database", err)
		}
	})
}

// GetDB 返回数据库连接池对象
func GetMysqlDB() *sqlx.DB {
	if mysqldb == nil {
		fmt.Printf("Database not initialized. Call InitDB first.")
	}
	return mysqldb
}
