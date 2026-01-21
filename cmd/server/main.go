// @title           政府采购网监控系统 API
// @version         1.0
// @description     政府采购网公告监控系统的 API 文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  403608355@qq.com

// @host      localhost:8080
// @BasePath  /api

package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/ieasydevops/demo-scrapy/docs"
	"github.com/ieasydevops/demo-scrapy/internal/api"
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

	for _, wp := range cfg.WebPages {
		result, _ := database.DB.Exec("INSERT OR IGNORE INTO web_pages (url, name) VALUES (?, ?)", wp.URL, wp.Name)
		log.Printf("初始化网页配置: %s", wp.Name)
		_ = result
	}

	for _, keyword := range cfg.Keywords {
		database.DB.Exec("INSERT OR IGNORE INTO keywords (keyword) VALUES (?)", keyword)
		log.Printf("初始化关键词: %s", keyword)
	}

	database.DB.Exec("INSERT OR IGNORE INTO push_config (email, push_time) VALUES (?, ?)",
		cfg.Email.SMTPUser, "17")

	for _, mc := range cfg.MonitorConfigs {
		var webPageID int64
		err := database.DB.QueryRow("SELECT id FROM web_pages WHERE name = ?", mc.WebPageName).Scan(&webPageID)
		if err == nil && webPageID > 0 {
			keywordsStr := strings.Join(mc.Keywords, ",")
			database.DB.Exec(`INSERT OR IGNORE INTO monitor_config 
				(web_page_id, crawl_time, crawl_freq, keywords) VALUES (?, ?, ?, ?)`,
				webPageID, mc.CrawlTime, mc.CrawlFreq, keywordsStr)
			log.Printf("初始化监控配置: %s, 时间=%s, 频率=%s, 关键词=%v", mc.WebPageName, mc.CrawlTime, mc.CrawlFreq, mc.Keywords)
		}
	}

	scheduler.Start()
	if err := scheduler.ReloadTasks(); err != nil {
		log.Printf("加载定时任务失败: %v", err)
	}

	log.Println("启动后立即执行一次采集任务...")
	go func() {
		time.Sleep(2 * time.Second)
		scheduler.ExecuteCrawlTask()
	}()

	port := cfg.Server.Port
	if port == 0 {
		port = 8080
	}
	r := api.SetupRouter()
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
