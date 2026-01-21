package crawler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ieasydevops/demo-scrapy/internal/database"
	"github.com/ieasydevops/demo-scrapy/internal/models"
)

func CrawlPage(url string, keywords []string) ([]models.Announcement, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Referer", "http://zfcg.szggzy.com:8081/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var announcements []models.Announcement

	doc.Find("ul.news-items li").Each(func(i int, s *goquery.Selection) {
		titleElem := s.Find("a.text-overflow")
		if titleElem.Length() == 0 {
			titleElem = s.Find("a")
		}
		if titleElem.Length() == 0 {
			return
		}

		title := strings.TrimSpace(titleElem.Text())
		href, exists := titleElem.Attr("href")
		if !exists {
			return
		}

		dateElem := s.Find("span.news-time")
		publishDate := strings.TrimSpace(dateElem.Text())

		if len(keywords) > 0 {
			matched := false
			for _, keyword := range keywords {
				if strings.Contains(title, keyword) {
					matched = true
					break
				}
			}
			if !matched {
				return
			}
		}

		fullURL := href
		if !strings.HasPrefix(href, "http") {
			if strings.HasPrefix(href, "//") {
				fullURL = "http:" + href
			} else if strings.HasPrefix(href, "/") {
				fullURL = "http://zfcg.szggzy.com:8081" + href
			} else {
				fullURL = "http://zfcg.szggzy.com:8081/" + href
			}
		}

		announcement := models.Announcement{
			Title:       title,
			URL:         fullURL,
			PublishDate: publishDate,
		}

		announcements = append(announcements, announcement)
	})

	if len(announcements) == 0 {
		doc.Find("tr").Each(func(i int, s *goquery.Selection) {
			titleElem := s.Find("td a")
			if titleElem.Length() == 0 {
				return
			}

			title := strings.TrimSpace(titleElem.Text())
			href, exists := titleElem.Attr("href")
			if !exists {
				return
			}

			dateElem := s.Find("td").Last()
			publishDate := strings.TrimSpace(dateElem.Text())

			if len(keywords) > 0 {
				matched := false
				for _, keyword := range keywords {
					if strings.Contains(title, keyword) {
						matched = true
						break
					}
				}
				if !matched {
					return
				}
			}

			fullURL := href
			if !strings.HasPrefix(href, "http") {
				if strings.HasPrefix(href, "//") {
					fullURL = "http:" + href
				} else if strings.HasPrefix(href, "/") {
					fullURL = "http://zfcg.szggzy.com:8081" + href
				} else {
					fullURL = "http://zfcg.szggzy.com:8081/" + href
				}
			}

			announcement := models.Announcement{
				Title:       title,
				URL:         fullURL,
				PublishDate: publishDate,
			}

			announcements = append(announcements, announcement)
		})
	}

	return announcements, nil
}

func FetchAnnouncementDetail(detailURL string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", detailURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "http://zfcg.szggzy.com:8081/gsgg/secondPage.html")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	var contentParts []string

	doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find("td").First().Text())
		value := strings.TrimSpace(s.Find("td").Last().Text())
		if label != "" && value != "" {
			contentParts = append(contentParts, fmt.Sprintf("%s: %s", label, value))
		}
	})

	if len(contentParts) == 0 {
		doc.Find(".content, .detail-content, #content").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				contentParts = append(contentParts, text)
			}
		})
	}

	if len(contentParts) == 0 {
		contentParts = append(contentParts, doc.Find("body").Text())
	}

	return strings.Join(contentParts, "\n\n"), nil
}

func CrawlPageWithDetails(url string, keywords []string, fetchDetails bool) ([]models.Announcement, error) {
	announcements, err := CrawlPage(url, keywords)
	if err != nil {
		return nil, err
	}

	if fetchDetails {
		for i := range announcements {
			content, err := FetchAnnouncementDetail(announcements[i].URL)
			if err == nil {
				announcements[i].Content = content
			}
			time.Sleep(200 * time.Millisecond)
		}
	}

	return announcements, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type ApiAnnouncement struct {
	ID                  string `json:"id"`
	PurchaseProjectName string `json:"purchaseProjectName"`
	Purchaser           string `json:"purchaser"`
	PublishDate         string `json:"publishDate"`
	URL                 string `json:"url"`
	PurchaseAgent       string `json:"purchaseAgent"`
}

type ApiResponse struct {
	Total int               `json:"total"`
	Data  []ApiAnnouncement `json:"data"`
	Code  int               `json:"code"`
	Msg   string            `json:"msg"`
}

func CrawlPageByAPI(pageURL string, keywords []string) ([]models.Announcement, error) {
	apiURLs := []string{
		"http://zfcg.szggzy.com:8081/gsgg/querySecondPageGsgg",
		"http://zfcg.szggzy.com:8081/gsgg/querySecondPageGsgg.do",
		"http://zfcg.szggzy.com:8081/gsgg/querySecondPageGsgg.json",
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	var allAnnouncements []models.Announcement

	req, err := http.NewRequest("GET", pageURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		client.Do(req)
	}

	time.Sleep(500 * time.Millisecond)

	var apiURL string
	var foundValidURL bool

	for _, testURL := range apiURLs {
		formData := url.Values{}
		formData.Set("pageIndex", "1")
		formData.Set("pageSize", "1")
		formData.Set("xxlx", "")

		req, err := http.NewRequest("POST", testURL, strings.NewReader(formData.Encode()))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Origin", "http://zfcg.szggzy.com:8081")
		req.Header.Set("Referer", "http://zfcg.szggzy.com:8081/gsgg/secondPage.html")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if !strings.HasPrefix(strings.TrimSpace(string(body)), "<") {
				apiURL = testURL
				foundValidURL = true
				break
			}
		}
		resp.Body.Close()
	}

	if !foundValidURL {
		return nil, fmt.Errorf("无法找到有效的 API 端点")
	}

	for page := 1; page <= 10; page++ {
		formData := url.Values{}
		formData.Set("pageIndex", fmt.Sprintf("%d", page))
		formData.Set("pageSize", "20")
		formData.Set("xxlx", "")

		req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Origin", "http://zfcg.szggzy.com:8081")
		req.Header.Set("Referer", "http://zfcg.szggzy.com:8081/gsgg/secondPage.html")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)
			if page == 1 {
				return nil, fmt.Errorf("HTTP状态码错误: %d，响应内容: %s", resp.StatusCode, bodyStr[:min(300, len(bodyStr))])
			}
			break
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}

		bodyStr := string(body)
		if strings.HasPrefix(strings.TrimSpace(bodyStr), "<") {
			return nil, fmt.Errorf("服务器返回HTML而非JSON，可能是请求格式错误。响应前100字符: %s", bodyStr[:min(100, len(bodyStr))])
		}

		var apiResp ApiResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("JSON解析失败: %v，响应内容: %s", err, bodyStr[:min(200, len(bodyStr))])
		}

		if apiResp.Code != 0 && apiResp.Code != 200 {
			if page == 1 {
				return nil, fmt.Errorf("API返回错误: %s", apiResp.Msg)
			}
			break
		}

		if len(apiResp.Data) == 0 {
			break
		}

		for _, item := range apiResp.Data {
			if len(keywords) > 0 {
				matched := false
				for _, keyword := range keywords {
					if strings.Contains(item.Purchaser, keyword) || strings.Contains(item.PurchaseProjectName, keyword) {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}

			fullURL := item.URL
			if fullURL == "" {
				if item.ID != "" {
					fullURL = fmt.Sprintf("http://zfcg.szggzy.com:8081/gsgg/detail/%s.html", item.ID)
				}
			} else if !strings.HasPrefix(fullURL, "http") {
				if strings.HasPrefix(fullURL, "//") {
					fullURL = "http:" + fullURL
				} else if strings.HasPrefix(fullURL, "/") {
					fullURL = "http://zfcg.szggzy.com:8081" + fullURL
				} else {
					fullURL = "http://zfcg.szggzy.com:8081/" + fullURL
				}
			}

			announcement := models.Announcement{
				Title:       item.PurchaseProjectName,
				URL:         fullURL,
				PublishDate: item.PublishDate,
			}

			allAnnouncements = append(allAnnouncements, announcement)
		}

		if len(apiResp.Data) < 20 {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	return allAnnouncements, nil
}

func SaveAnnouncements(announcements []models.Announcement, webPageID int) error {
	if len(announcements) == 0 {
		return nil
	}

	savedCount := 0
	skippedCount := 0

	for _, ann := range announcements {
		var existingID int
		var existingPublisher string
		err := database.DB.QueryRow("SELECT id, COALESCE(publisher, '') FROM announcements WHERE url = ?", ann.URL).Scan(&existingID, &existingPublisher)

		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				_, err = database.DB.Exec(
					"INSERT INTO announcements (title, url, publish_date, content, web_page_id, publisher) VALUES (?, ?, ?, ?, ?, ?)",
					ann.Title, ann.URL, ann.PublishDate, ann.Content, webPageID, ann.Publisher,
				)
				if err != nil {
					return fmt.Errorf("插入公告失败: %v, URL: %s", err, ann.URL)
				}
				savedCount++
			} else {
				return fmt.Errorf("检查公告是否存在失败: %v", err)
			}
		} else {
			if ann.Publisher != "" && (existingPublisher == "" || existingPublisher != ann.Publisher) {
				_, err = database.DB.Exec(
					"UPDATE announcements SET title = ?, publish_date = ?, content = ?, publisher = ? WHERE id = ?",
					ann.Title, ann.PublishDate, ann.Content, ann.Publisher, existingID,
				)
				if err != nil {
					log.Printf("更新公告失败: %v, ID: %d", err, existingID)
				} else {
					savedCount++
					log.Printf("更新公告采购单位: ID=%d, Publisher=%s", existingID, ann.Publisher)
				}
			} else {
				skippedCount++
			}
		}
	}

	log.Printf("保存公告完成: 新增 %d 条, 跳过 %d 条(已存在)", savedCount, skippedCount)
	return nil
}

func GetNewAnnouncements() ([]models.Announcement, error) {
	rows, err := database.DB.Query(`
		SELECT a.id, a.title, a.url, a.publish_date, a.content, a.created_at,
		       a.web_page_id, wp.name as web_page_name, a.publisher
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
			&ann.CreatedAt, &ann.WebPageID, &ann.WebPageName, &ann.Publisher)
		if err != nil {
			return nil, err
		}
		announcements = append(announcements, ann)
	}

	return announcements, nil
}

type APISearchRequest struct {
	Token          string      `json:"token"`
	Pn             int         `json:"pn"`
	Rn             int         `json:"rn"`
	Sdt            string      `json:"sdt"`
	Edt            string      `json:"edt"`
	Wd             string      `json:"wd"`
	IncWd          string      `json:"inc_wd"`
	ExcWd          string      `json:"exc_wd"`
	Fields         string      `json:"fields"`
	Cnum           string      `json:"cnum"`
	Sort           string      `json:"sort"`
	Ssort          string      `json:"ssort"`
	Cl             int         `json:"cl"`
	Terminal       string      `json:"terminal"`
	Condition      interface{} `json:"condition"`
	Time           interface{} `json:"time"`
	Highlights     string      `json:"highlights"`
	Statistics     interface{} `json:"statistics"`
	UnionCondition interface{} `json:"unionCondition"`
	Accuracy       string      `json:"accuracy"`
	NoParticiple   string      `json:"noParticiple"`
	SearchRange    interface{} `json:"searchRange"`
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

			publisher, err := extractPublisherFromDetailURL(client, fullURL)
			if err != nil || publisher == "" || publisher == "未知单位" {
				log.Printf("从详情页提取采购单位失败，尝试从内容提取: %v", err)
				publisher = extractPublisherFromContentText(record.Content)
			}

			if publisher == "" || publisher == "未知单位" || len(publisher) > 100 {
				log.Printf("未能提取到有效采购单位，标题: %s, 提取值: %s", cleanTitle, publisher)
				publisher = ""
			}

			matched := false
			if len(keywords) > 0 {
				for _, keyword := range keywords {
					if strings.Contains(publisher, keyword) || strings.Contains(cleanTitle, keyword) {
						matched = true
						break
					}
				}
			} else {
				if publisher != "" {
					matched = strings.Contains(publisher, "生态环境局")
				} else {
					matched = strings.Contains(cleanTitle, "生态环境局")
				}
			}

			if matched {
				announcement := models.Announcement{
					Title:       cleanTitle,
					URL:         fullURL,
					PublishDate: formatDateString(record.Webdate),
					Content:     cleanHTMLTags(record.Content),
					Publisher:   publisher,
				}
				allAnnouncements = append(allAnnouncements, announcement)
				if publisher != "" {
					log.Printf("采集公告: %s, 采购单位: %s", cleanTitle, publisher)
				} else {
					log.Printf("采集公告: %s (采购单位待补充)", cleanTitle)
				}
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

func extractPublisherFromDetailURL(client *http.Client, urlStr string) (string, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "http://zfcg.szggzy.com:8081/")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	publisher := extractPublisherFromHTMLText(string(body))
	if publisher == "" || publisher == "未知单位" {
		return "", fmt.Errorf("未能提取到采购单位")
	}

	return publisher, nil
}

func extractPublisherFromHTMLText(htmlContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err == nil {
		var publisher string
		doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
			label := strings.TrimSpace(s.Find("td").First().Text())
			if strings.Contains(label, "采购单位") || strings.Contains(label, "采购人") {
				value := strings.TrimSpace(s.Find("td").Last().Text())
				if value != "" && publisher == "" {
					value = strings.ReplaceAll(value, "\n", " ")
					value = strings.ReplaceAll(value, "\r", " ")
					value = strings.ReplaceAll(value, "\t", " ")
					value = regexp.MustCompile(`\s+`).ReplaceAllString(value, " ")
					value = strings.TrimSpace(value)

					if len(value) > 0 && len(value) <= 100 && !strings.Contains(value, "项目") && !strings.Contains(value, "合同") {
						publisher = value
					}
				}
			}
		})
		doc.Find("p, div, span").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if strings.Contains(text, "采购人（甲方）") || strings.Contains(text, "采购人(甲方)") {
				re := regexp.MustCompile(`采购人[（(]甲方[）)]\s*[：:]\s*([^<&\n\r]+?)(?:地址|联系方式|采购代理|项目联系|</|$)`)
				matches := re.FindStringSubmatch(text)
				if len(matches) > 1 && publisher == "" {
					value := strings.TrimSpace(matches[1])
					value = strings.ReplaceAll(value, "\n", " ")
					value = strings.ReplaceAll(value, "\r", " ")
					value = strings.ReplaceAll(value, "\t", " ")
					value = regexp.MustCompile(`\s+`).ReplaceAllString(value, " ")
					value = strings.TrimSpace(value)
					if len(value) > 0 && len(value) <= 100 && !strings.Contains(value, "项目") && !strings.Contains(value, "合同") {
						publisher = value
					}
				}
			}
		})
		if publisher != "" {
			log.Printf("从表格提取到采购单位: %s", publisher)
			return publisher
		}
	}

	patterns := []string{
		`采购人[（(]甲方[）)]\s*[：:]\s*([^<&\n\r]+?)(?:地址|联系方式|采购代理|项目联系|</|$)`,
		`采购单位[：:]\s*([^<&\n\r]+?)(?:地址|联系方式|采购代理|项目联系|</)`,
		`采购人[信息]?[：:]\s*(?:名称[：:]?\s*)?([^<&\n\r]+?)(?:地址|联系方式|采购代理|项目联系|</)`,
		`采购人名称[：:]\s*([^<&\n\r]+?)(?:地址|联系方式|</)`,
		`<td[^>]*>\s*采购单位\s*</td>\s*<td[^>]*>\s*([^<]+?)\s*</td>`,
		`<td[^>]*>\s*(?:采购人|招标人)\s*</td>\s*<td[^>]*>\s*([^<]+?)\s*</td>`,
		`<td[^>]*>\s*(?:采购人|招标人)名称\s*</td>\s*<td[^>]*>\s*([^<]+?)\s*</td>`,
		`<td[^>]*>\s*采购人[（(]甲方[）)]\s*</td>\s*<td[^>]*>\s*([^<]+?)\s*</td>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			publisher := strings.TrimSpace(matches[1])
			publisher = strings.ReplaceAll(publisher, "&nbsp;", " ")
			publisher = strings.ReplaceAll(publisher, "&#160;", " ")
			publisher = strings.ReplaceAll(publisher, "&amp;", "&")
			publisher = strings.ReplaceAll(publisher, "&lt;", "<")
			publisher = strings.ReplaceAll(publisher, "&gt;", ">")
			publisher = strings.ReplaceAll(publisher, "\n", " ")
			publisher = strings.ReplaceAll(publisher, "\r", " ")
			publisher = strings.ReplaceAll(publisher, "\t", " ")
			publisher = regexp.MustCompile(`\s+`).ReplaceAllString(publisher, " ")
			publisher = strings.TrimSpace(publisher)
			if publisher != "" && len(publisher) > 0 && len(publisher) <= 100 && !strings.Contains(publisher, "项目") && !strings.Contains(publisher, "合同") {
				log.Printf("从正则表达式提取到采购单位: %s (模式: %s)", publisher, pattern)
				return publisher
			}
		}
	}

	return extractPublisherFromContentText(htmlContent)
}

func extractPublisherFromContentText(content string) string {
	cleanContent := cleanHTMLTags(content)
	sentences := strings.Split(cleanContent, "。")
	for _, sentence := range sentences {
		if strings.Contains(sentence, "生态环境局") {
			runes := []rune(sentence)
			keywordRunes := []rune("生态环境局")

			idxRune := -1
			for i := 0; i <= len(runes)-len(keywordRunes); i++ {
				match := true
				for j := 0; j < len(keywordRunes); j++ {
					if runes[i+j] != keywordRunes[j] {
						match = false
						break
					}
				}
				if match {
					idxRune = i
					break
				}
			}

			if idxRune < 0 {
				continue
			}

			start := idxRune - 1
			for start >= 0 && start < len(runes) && !isChinesePunctuationRune(runes[start]) {
				start--
			}
			if start < 0 {
				start = 0
			} else {
				start++
			}

			end := idxRune + len(keywordRunes)
			for end < len(runes) && !isChinesePunctuationRune(runes[end]) {
				end++
			}

			if start < len(runes) && end <= len(runes) && start < end {
				return strings.TrimSpace(string(runes[start:end]))
			}
		}
	}
	return "未知单位"
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

func isChinesePunctuationRune(c rune) bool {
	punctuations := []rune{'，', '。', '；', '：', '？', '！', '、', '（', '）', '《', '》', '【', '】'}
	for _, p := range punctuations {
		if c == p {
			return true
		}
	}
	return false
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
