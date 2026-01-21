package crawler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ieasydevops/demo-scrapy/internal/database"
	"github.com/ieasydevops/demo-scrapy/internal/models"
)

type APISearchRequest struct {
	Pn           int    `json:"pn"`
	Rn           int    `json:"rn"`
	Sdt          string `json:"sdt"`
	Edt          string `json:"edt"`
	Wd           string `json:"wd"`
	Fields       string `json:"fields"`
	Cnum         string `json:"cnum"`
	Sort         string `json:"sort"`
	Ssort        string `json:"ssort"`
	Cl           int    `json:"cl"`
	Highlights   string `json:"highlights"`
	NoParticiple string `json:"noParticiple"`
}

type APISearchResponse struct {
	Result struct {
		Totalcount int `json:"totalcount"`
		Records    []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Webdate string `json:"webdate"`
			Linkurl string `json:"linkurl"`
		} `json:"records"`
	} `json:"result"`
}

func CrawlByAPISearch(keywords []string, days int) ([]models.Announcement, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	keywordStr := strings.Join(keywords, " ")
	if keywordStr == "" {
		keywordStr = "生态环境局"
	}

	var allAnnouncements []models.Announcement
	pageNum := 0
	pageSize := 50

	for {
		searchReq := APISearchRequest{
			Pn:           pageNum * pageSize,
			Rn:           pageSize,
			Sdt:          startTime.Format("2006-01-02 15:04:05"),
			Edt:          endTime.Format("2006-01-02 15:04:05"),
			Wd:           keywordStr,
			Fields:       "title;content",
			Cnum:         "002",
			Sort:         "{\"webdate\":\"0\"}",
			Ssort:        "title",
			Cl:           500,
			Highlights:   "title;content",
			NoParticiple: "0",
		}

		apiResponse, err := sendAPISearchRequest(client, searchReq)
		if err != nil {
			return nil, fmt.Errorf("API请求失败: %v", err)
		}

		if apiResponse.Result.Totalcount == 0 {
			break
		}

		for _, record := range apiResponse.Result.Records {
			cleanTitle := cleanHTMLTags(record.Title)
			fullURL := buildFullURLFromAPI(record.Linkurl)

			matched := false
			if len(keywords) > 0 {
				for _, keyword := range keywords {
					if strings.Contains(cleanTitle, keyword) || strings.Contains(cleanHTMLTags(record.Content), keyword) {
						matched = true
						break
					}
				}
			} else {
				matched = strings.Contains(cleanTitle, "生态环境局")
			}

			if matched {
				announcement := models.Announcement{
					Title:       cleanTitle,
					URL:         fullURL,
					PublishDate: formatDateString(record.Webdate),
					Content:     cleanHTMLTags(record.Content),
				}
				allAnnouncements = append(allAnnouncements, announcement)
				log.Printf("采集公告: %s", cleanTitle)
			}
		}

		if pageNum*pageSize+len(apiResponse.Result.Records) >= apiResponse.Result.Totalcount {
			break
		}

		if len(apiResponse.Result.Records) == 0 {
			break
		}

		pageNum++
		time.Sleep(500 * time.Millisecond)
	}

	return allAnnouncements, nil
}

func sendAPISearchRequest(client *http.Client, reqData APISearchRequest) (*APISearchResponse, error) {
	apiURL := "http://zfcg.szggzy.com:8081/inteligentsearch/rest/esinteligentsearch/getFullTextDataNew"

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("JSON编码失败: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "http://zfcg.szggzy.com:8081")
	req.Header.Set("Referer", "http://zfcg.szggzy.com:8081/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var apiResp APISearchResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}

	return &apiResp, nil
}

func SaveAnnouncements(announcements []models.Announcement, webPageID int) error {
	if len(announcements) == 0 {
		return nil
	}

	savedCount := 0
	skippedCount := 0

	for _, ann := range announcements {
		var existingID int
		err := database.DB.QueryRow("SELECT id FROM announcements WHERE url = ?", ann.URL).Scan(&existingID)

		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				_, err = database.DB.Exec(
					"INSERT INTO announcements (title, url, publish_date, content, web_page_id) VALUES (?, ?, ?, ?, ?)",
					ann.Title, ann.URL, ann.PublishDate, ann.Content, webPageID,
				)
				if err != nil {
					return fmt.Errorf("插入公告失败: %v, URL: %s", err, ann.URL)
				}
				savedCount++
			} else {
				return fmt.Errorf("检查公告是否存在失败: %v", err)
			}
		} else {
			skippedCount++
		}
	}

	log.Printf("保存公告完成: 新增 %d 条, 跳过 %d 条(已存在)", savedCount, skippedCount)
	return nil
}

func GetNewAnnouncements() ([]models.Announcement, error) {
	rows, err := database.DB.Query(`
		SELECT a.id, a.title, a.url, a.publish_date, a.content, a.created_at,
		       a.web_page_id, wp.name as web_page_name
		FROM announcements a
		LEFT JOIN web_pages wp ON a.web_page_id = wp.id
		WHERE DATE(a.created_at) = DATE('now')
		ORDER BY a.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announcements []models.Announcement
	for rows.Next() {
		var ann models.Announcement
		err := rows.Scan(&ann.ID, &ann.Title, &ann.URL, &ann.PublishDate, &ann.Content,
			&ann.CreatedAt, &ann.WebPageID, &ann.WebPageName)
		if err != nil {
			return nil, err
		}
		announcements = append(announcements, ann)
	}

	return announcements, nil
}

func cleanHTMLTags(html string) string {
	html = strings.ReplaceAll(html, "<em style='color:red'>", "")
	html = strings.ReplaceAll(html, "</em>", "")
	re := regexp.MustCompile(`<[^>]*>`)
	html = re.ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	html = strings.ReplaceAll(html, "&#160;", " ")
	html = strings.ReplaceAll(html, "&amp;", "&")
	return strings.TrimSpace(html)
}

func formatDateString(dateStr string) string {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return dateStr
}

func buildFullURLFromAPI(href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "/") {
		return "http://zfcg.szggzy.com:8081" + href
	}
	return "http://zfcg.szggzy.com:8081/gsgg/" + href
}
