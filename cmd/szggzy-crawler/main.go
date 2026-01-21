package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// 定义公告结构体
type Announcement struct {
	Title       string `json:"title"`
	Publisher   string `json:"publisher"`
	PublishDate string `json:"publishDate"`
	URL         string `json:"url"`
	Content     string `json:"content,omitempty"`
}

// 分页信息
type PageInfo struct {
	CurrentPage int
	TotalPages  int
	TotalItems  int
	PageSize    int
}

func main() {
	fmt.Println("开始采集深圳市公共资源交易中心生态环境局公告...")
	fmt.Println(strings.Repeat("=", 60))

	// 设置查询条件：最近一个月
	oneMonthAgo := time.Now().AddDate(0, -1, 0)
	fmt.Printf("采集时间范围: %s 至今\n", oneMonthAgo.Format("2006-01-02"))

	// 采集公告
	announcements, err := fetchEcologyAnnouncements(oneMonthAgo)
	if err != nil {
		fmt.Printf("采集失败: %v\n", err)
		return
	}

	// 输出结果
	if len(announcements) == 0 {
		fmt.Println("未找到生态环境局相关的公告")
	} else {
		fmt.Printf("共找到 %d 条生态环境局公告:\n", len(announcements))
		fmt.Println(strings.Repeat("=", 60))

		for i, ann := range announcements {
			fmt.Printf("\n公告 #%d:\n", i+1)
			fmt.Printf("标题: %s\n", ann.Title)
			fmt.Printf("发布单位: %s\n", ann.Publisher)
			fmt.Printf("发布日期: %s\n", ann.PublishDate)
			fmt.Printf("公告链接: %s\n", ann.URL)

			// 如果内容不为空且较短，可以显示部分内容
			if len(ann.Content) > 0 && len(ann.Content) < 200 {
				fmt.Printf("内容摘要: %s\n", ann.Content)
			}
			fmt.Println(strings.Repeat("-", 40))
		}
	}
}

// 获取生态环境局公告
func fetchEcologyAnnouncements(since time.Time) ([]Announcement, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var ecologyAnnouncements []Announcement
	currentPage := 1
	maxPages := 20 // 最大采集20页，防止无限循环

	fmt.Println("开始分页采集公告...")

	for currentPage <= maxPages {
		fmt.Printf("正在采集第 %d 页...\n", currentPage)

		// 获取当前页的公告列表
		announcements, err := fetchAnnouncementList(client, currentPage)
		if err != nil {
			fmt.Printf("第 %d 页采集失败: %v\n", currentPage, err)
			break
		}

		if len(announcements) == 0 {
			fmt.Printf("第 %d 页没有找到公告，停止采集\n", currentPage)
			break
		}

		fmt.Printf("第 %d 页找到 %d 条公告\n", currentPage, len(announcements))

		// 处理当前页的公告
		pageHasRecentAnnouncement := false
		for i, ann := range announcements {
			fmt.Printf("  处理第 %d 页的第 %d/%d 条公告: %s\n",
				currentPage, i+1, len(announcements), ann.Title)

			// 检查公告日期
			annDate, err := time.Parse("2006-01-02", ann.PublishDate)
			if err != nil {
				// 如果日期解析失败，跳过这条公告
				fmt.Printf("  警告: 日期解析失败: %s\n", ann.PublishDate)
				continue
			}

			// 如果公告日期早于查询起始时间，跳过
			if annDate.Before(since) {
				fmt.Printf("  跳过: 公告日期 %s 早于查询起始时间 %s\n",
					ann.PublishDate, since.Format("2006-01-02"))
				continue
			}

			pageHasRecentAnnouncement = true

			// 获取公告详情
			detailAnn, err := fetchAnnouncementDetail(client, ann.URL)
			if err != nil {
				fmt.Printf("  获取公告详情失败: %v\n", err)
				continue
			}

			// 检查是否包含生态环境局
			if isEcologyAnnouncement(detailAnn) {
				ecologyAnnouncements = append(ecologyAnnouncements, detailAnn)
				fmt.Printf("  找到生态环境局公告: %s (发布日期: %s)\n", detailAnn.Title, detailAnn.PublishDate)
			}

			// 避免请求过快，添加延迟
			time.Sleep(100 * time.Millisecond)
		}

		// 如果当前页没有最近的公告，停止采集
		if !pageHasRecentAnnouncement {
			fmt.Printf("第 %d 页没有最近一个月的公告，停止采集\n", currentPage)
			break
		}

		currentPage++

		// 避免请求过快
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("分页采集完成，共采集 %d 页\n", currentPage-1)
	return ecologyAnnouncements, nil
}

// 获取指定页的公告列表
func fetchAnnouncementList(client *http.Client, page int) ([]Announcement, error) {
	var listURL string

	// 根据页码构建URL（基于提供的JavaScript逻辑）
	if page == 1 {
		// 第一页
		listURL = "http://zfcg.szggzy.com:8081/gsgg/002001/002001002/list.html"
	} else if page <= 10 {
		// 第2-10页
		listURL = fmt.Sprintf("http://zfcg.szggzy.com:8081/gsgg/002001/002001002/%d.html", page)
	} else {
		// 第11页及以后
		listURL = fmt.Sprintf("http://zfcg.szggzy.com:8081/gsgg/002001/002001002/list.html?categoryNum=002001002&pageIndex=%d", page)
	}

	req, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	setCommonHeaders(req)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 使用goquery解析HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %v", err)
	}

	var announcements []Announcement

	// 查找公告列表 - 根据提供的HTML结构
	doc.Find("ul.news-items li").Each(func(i int, s *goquery.Selection) {
		// 提取链接和标题
		linkElem := s.Find("a.text-overflow")
		title := strings.TrimSpace(linkElem.Text())
		href, exists := linkElem.Attr("href")

		// 提取发布日期
		dateElem := s.Find("span.news-time")
		publishDate := strings.TrimSpace(dateElem.Text())

		if exists && title != "" && publishDate != "" {
			// 构建完整URL
			fullURL := buildFullURL(href)

			announcements = append(announcements, Announcement{
				Title:       title,
				PublishDate: publishDate,
				URL:         fullURL,
			})
		}
	})

	// 如果当前选择器找不到，尝试其他选择器
	if len(announcements) == 0 {
		doc.Find("li").Each(func(i int, s *goquery.Selection) {
			linkElem := s.Find("a")
			title := strings.TrimSpace(linkElem.Text())
			href, exists := linkElem.Attr("href")

			// 尝试多种日期选择器
			var publishDate string
			s.Find("span").Each(func(j int, span *goquery.Selection) {
				text := strings.TrimSpace(span.Text())
				// 检查是否像日期格式
				if strings.Contains(text, "202") && len(text) == 10 {
					publishDate = text
				}
			})

			if exists && title != "" && publishDate != "" {
				fullURL := buildFullURL(href)
				announcements = append(announcements, Announcement{
					Title:       title,
					PublishDate: publishDate,
					URL:         fullURL,
				})
			}
		})
	}

	return announcements, nil
}

// 获取公告详情
func fetchAnnouncementDetail(client *http.Client, urlStr string) (Announcement, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return Announcement{}, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	setCommonHeaders(req)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return Announcement{}, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Announcement{}, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Announcement{}, fmt.Errorf("读取响应失败: %v", err)
	}

	htmlContent := string(body)

	// 使用goquery解析详情页
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return Announcement{}, fmt.Errorf("解析HTML失败: %v", err)
	}

	// 提取标题 - 尝试多种选择器
	title := extractTitle(doc)

	// 提取发布单位
	publisher := extractPublisher(htmlContent)

	// 提取发布日期
	publishDate := extractPublishDate(doc, htmlContent)

	// 提取内容
	content := extractContent(doc)

	return Announcement{
		Title:       title,
		Publisher:   publisher,
		PublishDate: publishDate,
		URL:         urlStr,
		Content:     content,
	}, nil
}

// 设置通用请求头
func setCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "http://zfcg.szggzy.com:8081/")
	req.Header.Set("Connection", "keep-alive")
}

// 提取标题
func extractTitle(doc *goquery.Document) string {
	titleSelectors := []string{
		"title",
		"h1",
		".title",
		".content-title",
		".article-title",
		"#title",
		".head h1",
	}

	for _, selector := range titleSelectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			if text := strings.TrimSpace(s.Text()); text != "" {
				// 如果已经找到标题，直接返回
				return
			}
		})
	}

	// 如果没找到，返回空字符串
	return ""
}

// 提取发布单位
func extractPublisher(htmlContent string) string {
	// 正则表达式模式
	patterns := []string{
		`发布单位[：:]\s*([^<\n\r]+)`,
		`采购人[：:]\s*([^<\n\r]+)`,
		`采购单位[：:]\s*([^<\n\r]+)`,
		`招标人[：:]\s*([^<\n\r]+)`,
		`<td[^>]*>\s*发布单位\s*</td>\s*<td[^>]*>\s*([^<]+)\s*</td>`,
		`<td[^>]*>\s*采购人\s*</td>\s*<td[^>]*>\s*([^<]+)\s*</td>`,
		`<td[^>]*>\s*招标人\s*</td>\s*<td[^>]*>\s*([^<]+)\s*</td>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			publisher := strings.TrimSpace(matches[1])
			if publisher != "" {
				return publisher
			}
		}
	}

	// 使用goquery查找
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err == nil {
		doc.Find("td, div, span, p").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if strings.Contains(text, "发布单位") || strings.Contains(text, "采购人") || strings.Contains(text, "招标人") {
				// 尝试获取包含单位信息的相邻元素
				html, _ := s.Html()
				if strings.Contains(html, "生态环境局") {
					// 在HTML片段中查找
					if re := regexp.MustCompile(`生态环境局[^<]*`); re.MatchString(html) {
						if match := re.FindString(html); match != "" {
							// 已经找到，直接返回
							return
						}
					}
				}
			}
		})
	}

	return "未知单位"
}

// 提取发布日期
func extractPublishDate(doc *goquery.Document, htmlContent string) string {
	// 先尝试从meta标签提取
	doc.Find("meta[name='publishdate'], meta[name='PubDate'], meta[property='article:published_time']").Each(func(i int, s *goquery.Selection) {
		if date, exists := s.Attr("content"); exists && date != "" {
			// 已经找到，直接返回
			return
		}
	})

	// 尝试常见的选择器
	dateSelectors := []string{
		".publish-date",
		".date",
		".time",
		".news-time",
		"time",
		".info span",
	}

	for _, selector := range dateSelectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			// 检查是否像日期格式
			if strings.Contains(text, "202") && (strings.Contains(text, "-") || strings.Contains(text, "/")) {
				// 已经找到，直接返回
				return
			}
		})
	}

	// 使用正则表达式查找日期
	datePatterns := []string{
		`\d{4}-\d{2}-\d{2}`,
		`\d{4}/\d{2}/\d{2}`,
		`\d{4}年\d{2}月\d{2}日`,
	}

	for _, pattern := range datePatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(htmlContent); len(matches) > 0 {
			return matches[0]
		}
	}

	return ""
}

// 提取内容
func extractContent(doc *goquery.Document) string {
	contentSelectors := []string{
		".content",
		".article-content",
		".main-content",
		".news-content",
		"#content",
		"#newsContent",
	}

	for _, selector := range contentSelectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			if text := strings.TrimSpace(s.Text()); text != "" {
				// 已经找到，直接返回
				return
			}
		})
	}

	// 如果上述选择器没找到内容，尝试获取body的主要内容
	doc.Find("body").Each(func(i int, s *goquery.Selection) {
		// 移除脚本和样式
		s.Find("script, style, nav, header, footer").Remove()
		if text := strings.TrimSpace(s.Text()); text != "" {
			// 已经找到，直接返回
			return
		}
	})

	return ""
}

// 检查是否为生态环境局公告
func isEcologyAnnouncement(ann Announcement) bool {
	// 检查发布单位是否包含生态环境局
	if strings.Contains(ann.Publisher, "生态环境局") {
		return true
	}

	// 检查标题是否包含生态环境局
	if strings.Contains(ann.Title, "生态环境局") {
		return true
	}

	// 检查内容是否包含生态环境局
	if strings.Contains(ann.Content, "生态环境局") {
		return true
	}

	return false
}

// 构建完整URL
func buildFullURL(href string) string {
	// 如果已经是完整URL，直接返回
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	// 如果以/开头，添加域名
	if strings.HasPrefix(href, "/") {
		return "http://zfcg.szggzy.com:8081" + href
	}

	// 否则可能需要添加基础路径
	return "http://zfcg.szggzy.com:8081/gsgg/" + href
}

// 解析分页信息（备用）
func parsePageInfo(doc *goquery.Document) *PageInfo {
	info := &PageInfo{
		CurrentPage: 1,
		PageSize:    10,
	}

	// 尝试从分页组件中提取信息
	doc.Find(".pagination, .ewb-page2, .page-info").Each(func(i int, s *goquery.Selection) {
		text := s.Text()

		// 尝试提取总条数
		if re := regexp.MustCompile(`共\s*(\d+)\s*条`); re.MatchString(text) {
			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
				if total, err := strconv.Atoi(matches[1]); err == nil {
					info.TotalItems = total
				}
			}
		}

		// 尝试提取当前页
		if re := regexp.MustCompile(`当前页\s*[:：]?\s*(\d+)`); re.MatchString(text) {
			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
				if current, err := strconv.Atoi(matches[1]); err == nil {
					info.CurrentPage = current
				}
			}
		}
	})

	// 计算总页数
	if info.TotalItems > 0 && info.PageSize > 0 {
		info.TotalPages = (info.TotalItems + info.PageSize - 1) / info.PageSize
	}

	return info
}

// 构建带参数的URL（备用）
func buildURLWithParams(baseURL string, params map[string]string) string {
	if len(params) == 0 {
		return baseURL
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String()
}
