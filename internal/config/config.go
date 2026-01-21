package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	WebPages       []WebPageConfig     `yaml:"web_pages"`
	Keywords       []string            `yaml:"keywords"`
	MonitorConfigs []MonitorConfigItem `yaml:"monitor_configs"`
	Email          EmailConfig         `yaml:"email"`
	Server         ServerConfig        `yaml:"server"`
}

type WebPageConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type MonitorConfigItem struct {
	WebPageName string   `yaml:"web_page_name"`
	CrawlTime   string   `yaml:"crawl_time"`
	CrawlFreq   string   `yaml:"crawl_freq"`
	Keywords    []string `yaml:"keywords"`
}

type EmailConfig struct {
	SMTPHost string `yaml:"smtp_host"`
	SMTPUser string `yaml:"smtp_user"`
	SMTPPass string `yaml:"smtp_pass"`
}

type ServerConfig struct {
	Port   int    `yaml:"port"`
	DBPath string `yaml:"db_path"`
}

var GlobalConfig *Config

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	GlobalConfig = &config
	return &config, nil
}

func InitDefaultConfig(configPath string) error {
	if _, err := os.Stat(configPath); err == nil {
		return nil
	}

	defaultConfig := Config{
		WebPages: []WebPageConfig{
			{
				Name: "深圳政府采购网",
				URL:  "http://zfcg.szggzy.com:8081/gsgg/secondPage.html",
			},
		},
		Keywords: []string{"生态环境局"},
		MonitorConfigs: []MonitorConfigItem{
			{
				WebPageName: "深圳政府采购网",
				CrawlTime:   "9",
				CrawlFreq:   "daily",
				Keywords:    []string{"生态环境局"},
			},
		},
		Email: EmailConfig{
			SMTPHost: "smtp.qq.com",
			SMTPUser: "403608355@qq.com",
			SMTPPass: "your_smtp_password",
		},
		Server: ServerConfig{
			Port:   5080,
			DBPath: "./monitor.db",
		},
	}

	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return fmt.Errorf("生成默认配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}
