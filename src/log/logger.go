package mlog

import (
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	// 创建一个新的日志记录器
	Logger = logrus.New()

	// 设置日志级别
	Logger.SetLevel(logrus.InfoLevel) // 你可以选择 logrus.InfoLevel, logrus.WarnLevel 等

	// 设置日志格式：JSON 格式（也可以使用 TextFormatter）
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,                  // 打印完整时间戳
		TimestampFormat: "2006-01-02 15:04:05", // 自定义时间格式
	})
}
