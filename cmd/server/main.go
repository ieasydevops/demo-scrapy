package main

import (
	"flag"
	"log"
	"time"

	"github.com/ieasydevops/demo-scrapy/internal/config"
	"github.com/ieasydevops/demo-scrapy/internal/database"
	"github.com/ieasydevops/demo-scrapy/internal/scheduler"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	if err := config.InitDefaultConfig(*configPath); err != nil {
		log.Printf("初始化默认配置失败: %v", err)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	if err := database.InitDB(cfg.Server.DBPath); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	database.DB.Exec("INSERT OR IGNORE INTO web_pages (url, name) VALUES (?, ?)",
		"http://zfcg.szggzy.com:8081/gsgg/secondPage.html", "深圳政府采购网")

	for _, keyword := range cfg.Keywords {
		database.DB.Exec("INSERT OR IGNORE INTO keywords (keyword) VALUES (?)", keyword)
		log.Printf("初始化关键词: %s", keyword)
	}

	database.DB.Exec("INSERT OR IGNORE INTO push_config (email, push_time) VALUES (?, ?)",
		cfg.Email.SMTPUser, "17")

	scheduler.Start()
	if err := scheduler.ReloadTasks(); err != nil {
		log.Printf("加载定时任务失败: %v", err)
	}

	log.Println("启动后立即执行一次采集任务...")
	go func() {
		time.Sleep(2 * time.Second)
		scheduler.ExecuteCrawlTask()
	}()

	log.Println("服务运行中，按 Ctrl+C 停止服务")
	select {}
}
