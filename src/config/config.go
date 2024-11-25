package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

var ConfigCache *Config

func init() {
	// 加载配置文件
	lodConfig, err := LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("加载配置文件异常: %v\n", err)
		return
	}
	ConfigCache = lodConfig
}

type Config struct {
	Mysql struct {
		User            string `yaml:"user"`
		Password        string `yaml:"pwd"`
		Host            string `yaml:"host"`
		DBName          string `yaml:"db"`
		MaxOpenConns    int    `yaml:"maxOpenConns"`
		MaxIdleConns    int    `yaml:"maxIdleConns"`
		ConnMaxLifetime int    `yaml:"connMaxLifetime"`
	} `yaml:"mysql"`

	Local struct {
		Number        string `yaml:"number"`
		ConfigItemUrl string `yaml:"configItemUrl"`
	} `yaml:"local"`

	FilePaths map[string]string `yaml:"filePaths"`
}

func LoadConfig(file string) (*Config, error) {
	// 打开文件
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 创建 Config 实例
	var cfg Config

	// 使用 YAML 解码
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
