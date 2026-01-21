package scheduler

import (
	"fmt"
	"log"

	"github.com/ieasydevops/demo-scrapy/internal/crawler"
	"github.com/ieasydevops/demo-scrapy/internal/database"
	"github.com/ieasydevops/demo-scrapy/internal/email"
	"github.com/robfig/cron/v3"
)

var c *cron.Cron

func Start() {
	c = cron.New()
	c.Start()
}

func ExecuteCrawlTask() {
	log.Println("开始执行采集任务...")

	var keywords []string
	keywordRows, err := database.DB.Query("SELECT keyword FROM keywords")
	if err == nil {
		defer keywordRows.Close()
		for keywordRows.Next() {
			var keyword string
			if err := keywordRows.Scan(&keyword); err == nil {
				keywords = append(keywords, keyword)
			}
		}
	}

	if len(keywords) == 0 {
		keywords = []string{"生态环境局"}
	}

	log.Printf("使用关键词进行API采集: %v", keywords)

	announcements, err := crawler.CrawlByAPISearch(keywords, 1)
	if err != nil {
		log.Printf("采集失败: %v", err)
		return
	}

	var webPageID int = 1
	err = database.DB.QueryRow("SELECT id FROM web_pages LIMIT 1").Scan(&webPageID)
	if err != nil {
		database.DB.Exec("INSERT OR IGNORE INTO web_pages (url, name) VALUES (?, ?)",
			"http://zfcg.szggzy.com:8081/gsgg/secondPage.html", "深圳政府采购网")
		webPageID = 1
	}

	if err := crawler.SaveAnnouncements(announcements, webPageID); err != nil {
		log.Printf("保存公告失败: %v", err)
		return
	}

	log.Printf("成功采集，获取 %d 条公告", len(announcements))
}

func ReloadTasks() error {
	c.Stop()
	c = cron.New()
	c.Start()

	_, err := c.AddFunc("*/10 * * * *", func() {
		log.Println("执行定时采集任务（每10分钟）...")
		ExecuteCrawlTask()
	})
	if err != nil {
		log.Printf("添加定时采集任务失败: %v", err)
	}

	var pushTime string = "17"
	var emailAddr string = "403608355@qq.com"

	err = database.DB.QueryRow("SELECT email, push_time FROM push_config LIMIT 1").Scan(&emailAddr, &pushTime)
	if err != nil {
		database.DB.Exec("INSERT OR IGNORE INTO push_config (email, push_time) VALUES (?, ?)", emailAddr, pushTime)
	}

	hour := pushTime
	if len(hour) == 1 {
		hour = "0" + hour
	}
	spec := fmt.Sprintf("0 %s * * *", hour)

	_, err = c.AddFunc(spec, func() {
		log.Println("执行定时邮件推送任务...")
		newAnnouncements, err := crawler.GetNewAnnouncements()
		if err != nil {
			log.Printf("获取新公告失败: %v", err)
			return
		}

		if len(newAnnouncements) > 0 {
			if err := email.SendEmail(emailAddr, newAnnouncements); err != nil {
				log.Printf("发送邮件失败: %v", err)
			} else {
				log.Printf("成功发送 %d 条公告到 %s", len(newAnnouncements), emailAddr)
			}
		}
	})

	return err
}
