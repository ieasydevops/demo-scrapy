package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ieasydevops/demo-scrapy/internal/crawler"
	"github.com/ieasydevops/demo-scrapy/internal/database"
	"github.com/ieasydevops/demo-scrapy/internal/models"
)

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	rows, err := database.DB.Query(`
		SELECT mc.id, mc.web_page_id, mc.keywords,
		       wp.url, wp.name
		FROM monitor_config mc
		LEFT JOIN web_pages wp ON mc.web_page_id = wp.id
	`)
	if err != nil {
		log.Fatalf("查询监控配置失败: %v", err)
	}
	defer rows.Close()

	type TaskConfig struct {
		ID        int
		WebPageID int
		URL       string
		Name      string
		Keywords  []string
	}

	var taskConfigs []TaskConfig
	for rows.Next() {
		var tc TaskConfig
		var keywordsStr string
		if err := rows.Scan(&tc.ID, &tc.WebPageID, &keywordsStr, &tc.URL, &tc.Name); err != nil {
			log.Printf("扫描配置失败: %v", err)
			continue
		}
		tc.Keywords = strings.Split(keywordsStr, ",")
		taskConfigs = append(taskConfigs, tc)
	}

	if len(taskConfigs) == 0 {
		log.Println("没有找到监控配置，尝试从 web_pages 表读取网页列表...")

		rows, err := database.DB.Query("SELECT id, url, name FROM web_pages")
		if err != nil {
			log.Fatalf("查询网页列表失败: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var tc TaskConfig
			if err := rows.Scan(&tc.WebPageID, &tc.URL, &tc.Name); err != nil {
				continue
			}

			keywordRows, err := database.DB.Query("SELECT keyword FROM keywords")
			if err == nil {
				var keywords []string
				for keywordRows.Next() {
					var keyword string
					if err := keywordRows.Scan(&keyword); err == nil {
						keywords = append(keywords, keyword)
					}
				}
				keywordRows.Close()
				tc.Keywords = keywords
			}

			taskConfigs = append(taskConfigs, tc)
		}
	}

	if len(taskConfigs) == 0 {
		log.Fatal("没有找到任何网页配置")
	}

	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("集成测试工具 - 网页信息采集验证")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println()

	totalAnnouncements := 0
	for i, tc := range taskConfigs {
		fmt.Printf("[%d/%d] 开始采集: %s\n", i+1, len(taskConfigs), tc.Name)
		fmt.Printf("URL: %s\n", tc.URL)
		fmt.Printf("关键词: %s\n", strings.Join(tc.Keywords, ", "))
		fmt.Println(strings.Repeat("-", 82))

		var announcements []models.Announcement
		var err error

		fetchDetails := os.Getenv("FETCH_DETAILS") == "true"

		if strings.Contains(tc.URL, "secondPage.html") {
			fmt.Println("检测到 secondPage.html，尝试使用 API 方式采集...")
			announcements, err = crawler.CrawlPageByAPI(tc.URL, tc.Keywords)
			if err != nil {
				fmt.Printf("⚠️  API 采集失败，回退到 HTML 解析方式: %v\n", err)
				announcements, err = crawler.CrawlPageWithDetails(tc.URL, tc.Keywords, fetchDetails)
			} else if fetchDetails {
				for i := range announcements {
					content, detailErr := crawler.FetchAnnouncementDetail(announcements[i].URL)
					if detailErr == nil {
						announcements[i].Content = content
					}
					time.Sleep(200 * time.Millisecond)
				}
			}
		} else {
			announcements, err = crawler.CrawlPageWithDetails(tc.URL, tc.Keywords, fetchDetails)
		}

		if err != nil {
			fmt.Printf("❌ 采集失败: %v\n", err)
			fmt.Println()
			continue
		}

		fmt.Printf("✅ 采集成功，获取 %d 条公告\n", len(announcements))
		fmt.Println()

		if len(announcements) > 0 {
			fmt.Println("采集结果详情:")
			for j, ann := range announcements {
				fmt.Printf("\n  [%d] %s\n", j+1, ann.Title)
				fmt.Printf("      URL: %s\n", ann.URL)
				fmt.Printf("      发布日期: %s\n", ann.PublishDate)
				if ann.Content != "" {
					contentPreview := ann.Content
					if len(contentPreview) > 200 {
						contentPreview = contentPreview[:200] + "..."
					}
					fmt.Printf("      内容预览: %s\n", contentPreview)
				}
			}
			fmt.Println()
		}

		totalAnnouncements += len(announcements)
		fmt.Println(strings.Repeat("=", 82))
		fmt.Println()
	}

	fmt.Println("测试总结:")
	fmt.Printf("  总网页数: %d\n", len(taskConfigs))
	fmt.Printf("  总公告数: %d\n", totalAnnouncements)
	fmt.Println()

	outputJSON := os.Getenv("OUTPUT_JSON")
	if outputJSON == "true" {
		type Result struct {
			TotalPages         int                      `json:"total_pages"`
			TotalAnnouncements int                      `json:"total_announcements"`
			Results            []map[string]interface{} `json:"results"`
		}

		var results []map[string]interface{}
		for _, tc := range taskConfigs {
			var announcements []models.Announcement
			var err error

			if strings.Contains(tc.URL, "secondPage.html") {
				announcements, err = crawler.CrawlPageByAPI(tc.URL, tc.Keywords)
			} else {
				announcements, err = crawler.CrawlPage(tc.URL, tc.Keywords)
			}

			if err != nil {
				continue
			}

			result := map[string]interface{}{
				"web_page": map[string]interface{}{
					"id":   tc.WebPageID,
					"name": tc.Name,
					"url":  tc.URL,
				},
				"keywords":      tc.Keywords,
				"announcements": announcements,
				"count":         len(announcements),
			}
			results = append(results, result)
		}

		jsonData := Result{
			TotalPages:         len(taskConfigs),
			TotalAnnouncements: totalAnnouncements,
			Results:            results,
		}

		jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
		if err == nil {
			fmt.Println("JSON 输出:")
			fmt.Println(string(jsonBytes))
		}
	}
}
