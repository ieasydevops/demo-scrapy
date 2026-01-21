package scheduler

import (
	"fmt"
	"log"
	"strings"

	"github.com/ieasydevops/demo/internal/crawler"
	"github.com/ieasydevops/demo/internal/database"
	"github.com/ieasydevops/demo/internal/email"
	"github.com/ieasydevops/demo/internal/models"
	"github.com/robfig/cron/v3"
)

var c *cron.Cron

func Start() {
	c = cron.New()
	c.Start()
}

func ExecuteCrawlTask() {
	log.Println("开始执行采集任务...")

	rows, err := database.DB.Query(`
		SELECT mc.id, mc.web_page_id, mc.crawl_time, mc.crawl_freq, mc.keywords,
		       wp.url, wp.name
		FROM monitor_config mc
		LEFT JOIN web_pages wp ON mc.web_page_id = wp.id
	`)
	if err != nil {
		log.Printf("查询监控配置失败: %v", err)
		return
	}
	defer rows.Close()

	type TaskConfig struct {
		ID        int
		WebPageID int
		URL       string
		Name      string
		CrawlTime string
		CrawlFreq string
		Keywords  []string
	}

	var taskConfigs []TaskConfig
	for rows.Next() {
		var tc TaskConfig
		var keywordsStr string
		if err := rows.Scan(&tc.ID, &tc.WebPageID, &tc.CrawlTime, &tc.CrawlFreq, &keywordsStr, &tc.URL, &tc.Name); err != nil {
			continue
		}
		tc.Keywords = strings.Split(keywordsStr, ",")
		taskConfigs = append(taskConfigs, tc)
	}

	if len(taskConfigs) == 0 {
		log.Println("没有配置监控任务，尝试使用默认配置...")

		var defaultWebPageID int
		var defaultURL, defaultName string
		err := database.DB.QueryRow("SELECT id, url, name FROM web_pages LIMIT 1").Scan(&defaultWebPageID, &defaultURL, &defaultName)
		if err != nil {
			log.Println("没有找到网页配置，跳过采集")
			return
		}

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

		log.Printf("使用默认配置进行采集: %s, 关键词: %v", defaultName, keywords)

		var announcements []models.Announcement
		var err2 error

		if strings.Contains(defaultURL, "secondPage.html") || len(keywords) > 0 {
			log.Printf("使用 API 搜索方式采集，关键词: %v", keywords)
			announcements, err2 = crawler.CrawlByAPISearch(keywords, 1)
		} else {
			log.Printf("使用 HTML 解析方式采集")
			announcements, err2 = crawler.CrawlPage(defaultURL, keywords)
		}

		if err2 != nil {
			log.Printf("采集失败: %v", err2)
			return
		}

		if err := crawler.SaveAnnouncements(announcements, defaultWebPageID); err != nil {
			log.Printf("保存公告失败: %v", err)
			return
		}

		log.Printf("成功采集 %s，获取 %d 条公告", defaultName, len(announcements))
		return
	}

	for _, tc := range taskConfigs {
		log.Printf("开始采集: %s", tc.Name)

		var announcements []models.Announcement
		var err error

		if strings.Contains(tc.URL, "secondPage.html") || len(tc.Keywords) > 0 {
			log.Printf("使用 API 搜索方式采集，关键词: %v", tc.Keywords)
			announcements, err = crawler.CrawlByAPISearch(tc.Keywords, 1)
		} else {
			log.Printf("使用 HTML 解析方式采集")
			announcements, err = crawler.CrawlPage(tc.URL, tc.Keywords)
		}

		if err != nil {
			log.Printf("采集失败: %v", err)
			continue
		}

		if err := crawler.SaveAnnouncements(announcements, tc.WebPageID); err != nil {
			log.Printf("保存公告失败: %v", err)
			continue
		}

		log.Printf("成功采集 %s，获取 %d 条公告", tc.Name, len(announcements))
	}
}

func AddTask(pushTime string, webPages []models.WebPage, keywords []string, emailAddr string) error {
	hour := pushTime
	if len(hour) == 1 {
		hour = "0" + hour
	}
	spec := fmt.Sprintf("0 %s * * *", hour)

	_, err := c.AddFunc(spec, func() {
		log.Println("开始执行定时任务...")

		for _, page := range webPages {
			announcements, err := crawler.CrawlPage(page.URL, keywords)
			if err != nil {
				log.Printf("爬取页面 %s 失败: %v", page.URL, err)
				continue
			}

			if err := crawler.SaveAnnouncements(announcements, page.ID); err != nil {
				log.Printf("保存公告失败: %v", err)
				continue
			}
		}

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

	rows, err := database.DB.Query(`
		SELECT mc.id, mc.web_page_id, mc.crawl_time, mc.crawl_freq, mc.keywords,
		       wp.url, wp.name
		FROM monitor_config mc
		LEFT JOIN web_pages wp ON mc.web_page_id = wp.id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type TaskConfig struct {
		ID        int
		WebPageID int
		URL       string
		Name      string
		CrawlTime string
		CrawlFreq string
		Keywords  []string
	}

	var taskConfigs []TaskConfig
	for rows.Next() {
		var tc TaskConfig
		var keywordsStr string
		if err := rows.Scan(&tc.ID, &tc.WebPageID, &tc.CrawlTime, &tc.CrawlFreq, &keywordsStr, &tc.URL, &tc.Name); err != nil {
			continue
		}
		tc.Keywords = strings.Split(keywordsStr, ",")
		taskConfigs = append(taskConfigs, tc)
	}

	rows, err = database.DB.Query("SELECT email, push_time FROM subscribe_config")
	if err != nil {
		return err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var emailAddr, time string
		if err := rows.Scan(&emailAddr, &time); err != nil {
			continue
		}
		emails = append(emails, emailAddr)
	}

	if len(taskConfigs) == 0 {
		return nil
	}

	for _, tc := range taskConfigs {
		tc := tc
		hour := tc.CrawlTime
		if len(hour) == 1 {
			hour = "0" + hour
		}
		spec := fmt.Sprintf("0 %s * * *", hour)

		_, err := c.AddFunc(spec, func() {
			log.Printf("开始执行定时任务: %s", tc.Name)

			announcements, err := crawler.CrawlPage(tc.URL, tc.Keywords)
			if err != nil {
				log.Printf("爬取页面 %s 失败: %v", tc.URL, err)
				return
			}

			if err := crawler.SaveAnnouncements(announcements, tc.WebPageID); err != nil {
				log.Printf("保存公告失败: %v", err)
				return
			}

			if len(emails) > 0 {
				newAnnouncements, err := crawler.GetNewAnnouncements()
				if err != nil {
					log.Printf("获取新公告失败: %v", err)
					return
				}

				if len(newAnnouncements) > 0 {
					for _, emailAddr := range emails {
						if err := email.SendEmail(emailAddr, newAnnouncements); err != nil {
							log.Printf("发送邮件到 %s 失败: %v", emailAddr, err)
						} else {
							log.Printf("成功发送 %d 条公告到 %s", len(newAnnouncements), emailAddr)
						}
					}
				}
			}
		})

		if err != nil {
			log.Printf("添加定时任务失败: %v", err)
		}
	}

	return nil
}
