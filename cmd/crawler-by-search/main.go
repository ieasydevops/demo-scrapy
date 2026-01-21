package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// 定义公告结构体
type Announcement struct {
	Title       string `json:"title"`
	Publisher   string `json:"publisher"`
	PublishDate string `json:"publishDate"`
	URL         string `json:"url"`
	Content     string `json:"content,omitempty"`
	Source      string `json:"source,omitempty"`
}

// API请求结构体
type SearchRequest struct {
	Token          string      `json:"token"`
	Pn             int         `json:"pn"`             // 页码
	Rn             int         `json:"rn"`             // 每页条数
	Sdt            string      `json:"sdt"`            // 开始时间
	Edt            string      `json:"edt"`            // 结束时间
	Wd             string      `json:"wd"`             // 关键词
	IncWd          string      `json:"inc_wd"`         // 包含的关键词
	ExcWd          string      `json:"exc_wd"`         // 排除的关键词
	Fields         string      `json:"fields"`         // 搜索字段
	Cnum           string      `json:"cnum"`           // 分类编号
	Sort           string      `json:"sort"`           // 排序
	Ssort          string      `json:"ssort"`          // 次要排序
	Cl             int         `json:"cl"`             // 搜索结果数量
	Terminal       string      `json:"terminal"`       // 终端类型
	Condition      interface{} `json:"condition"`      // 条件
	Time           interface{} `json:"time"`           // 时间条件
	Highlights     string      `json:"highlights"`     // 高亮字段
	Statistics     interface{} `json:"statistics"`     // 统计信息
	UnionCondition interface{} `json:"unionCondition"` // 联合条件
	Accuracy       string      `json:"accuracy"`       // 精确度
	NoParticiple   string      `json:"noParticiple"`   // 是否分词
	SearchRange    interface{} `json:"searchRange"`    // 搜索范围
}

// API响应结构体
type SearchResponse struct {
	Result struct {
		Categorys []struct {
			Categorynum  string `json:"categorynum"`
			Count        string `json:"count"`
			Categoryname string `json:"categoryname"`
		} `json:"categorys"`
		Totalcount int `json:"totalcount"`
		Records    []struct {
			Categorynum   string `json:"categorynum"`
			Sysclicktimes int    `json:"sysclicktimes"`
			Title         string `json:"title"`
			Content       string `json:"content"`
			Webdate       string `json:"webdate"`
			Highlight     struct {
				Title   string `json:"title,omitempty"`
				Content string `json:"content,omitempty"`
			} `json:"highlight"`
			Score          interface{} `json:"score"`
			Syscategory    string      `json:"syscategory"`
			Syscollectguid string      `json:"syscollectguid"`
			ID             string      `json:"id"`
			Linkurl        string      `json:"linkurl"`
			Sysscore       string      `json:"sysscore"`
			Infodate       string      `json:"infodate"`
		} `json:"records"`
		ScorllId    string `json:"scorllId"`
		Executetime string `json:"executetime"`
	} `json:"result"`
}

func main() {
	fmt.Println("开始采集深圳市公共资源交易中心生态环境局公告...")
	fmt.Println(strings.Repeat("=", 60))

	// 设置查询时间范围：最近一天
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -1)

	fmt.Printf("采集时间范围: %s 至 %s\n",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"))

	// 采集公告
	announcements, err := fetchEcologyAnnouncementsByAPI(startTime, endTime)
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
			if ann.Source != "" {
				fmt.Printf("来源: %s\n", ann.Source)
			}
			fmt.Println(strings.Repeat("-", 40))
		}

		// 可选：保存结果到文件
		saveResultsToFile(announcements)
	}
}

// 通过API接口获取生态环境局公告
func fetchEcologyAnnouncementsByAPI(startTime, endTime time.Time) ([]Announcement, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var allAnnouncements []Announcement
	pageNum := 0
	pageSize := 50 // 每页获取50条

	fmt.Println("开始通过API接口搜索生态环境局公告...")

	for {
		fmt.Printf("正在获取第 %d 页数据...\n", pageNum+1)

		// 构建搜索请求
		searchReq := SearchRequest{
			Token:          "",
			Pn:             pageNum * pageSize, // 从0开始
			Rn:             pageSize,
			Sdt:            startTime.Format("2006-01-02 15:04:05"),
			Edt:            endTime.Format("2006-01-02 15:04:05"),
			Wd:             "生态环境局",               // 搜索关键词
			IncWd:          "",                    // 包含的关键词
			ExcWd:          "",                    // 排除的关键词
			Fields:         "title;content",       // 搜索字段
			Cnum:           "002",                 // 分类编号
			Sort:           "{\"webdate\":\"0\"}", // 按时间倒序
			Ssort:          "title",
			Cl:             500, // 搜索结果数量
			Terminal:       "",
			Condition:      nil,
			Time:           nil,
			Highlights:     "title;content",
			Statistics:     nil,
			UnionCondition: nil,
			Accuracy:       "",
			NoParticiple:   "0",
			SearchRange:    nil,
		}

		// 发送API请求
		apiResponse, err := sendSearchRequest(client, searchReq)
		if err != nil {
			fmt.Printf("API请求失败: %v\n", err)
			break
		}

		if apiResponse.Result.Totalcount == 0 {
			fmt.Println("未搜索到相关公告")
			break
		}

		fmt.Printf("API返回 %d 条记录，总计 %d 条\n",
			len(apiResponse.Result.Records), apiResponse.Result.Totalcount)

		// 处理当前页的公告
		currentPageAnnouncements := processAPIRecords(client, apiResponse.Result.Records)
		allAnnouncements = append(allAnnouncements, currentPageAnnouncements...)

		// 如果已经获取了所有记录，或者当前页记录数小于pageSize，则停止
		if pageNum*pageSize+len(apiResponse.Result.Records) >= apiResponse.Result.Totalcount {
			fmt.Printf("已获取所有 %d 条记录，停止采集\n", apiResponse.Result.Totalcount)
			break
		}

		// 判断是否需要继续获取下一页
		if len(apiResponse.Result.Records) == 0 {
			fmt.Println("当前页没有记录，停止采集")
			break
		}

		pageNum++

		// 避免请求过快
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("API搜索完成，共获取 %d 条符合条件的公告\n", len(allAnnouncements))
	return allAnnouncements, nil
}

// 发送搜索请求
func sendSearchRequest(client *http.Client, reqData SearchRequest) (*SearchResponse, error) {
	apiURL := "http://zfcg.szggzy.com:8081/inteligentsearch/rest/esinteligentsearch/getFullTextDataNew"

	// 将请求数据转换为JSON
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("JSON编码失败: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	setAPIHeaders(req)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON响应
	var apiResp SearchResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}

	return &apiResp, nil
}

// 处理API返回的记录
func processAPIRecords(client *http.Client, records []struct {
	Categorynum   string `json:"categorynum"`
	Sysclicktimes int    `json:"sysclicktimes"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	Webdate       string `json:"webdate"`
	Highlight     struct {
		Title   string `json:"title,omitempty"`
		Content string `json:"content,omitempty"`
	} `json:"highlight"`
	Score          interface{} `json:"score"`
	Syscategory    string      `json:"syscategory"`
	Syscollectguid string      `json:"syscollectguid"`
	ID             string      `json:"id"`
	Linkurl        string      `json:"linkurl"`
	Sysscore       string      `json:"sysscore"`
	Infodate       string      `json:"infodate"`
}) []Announcement {
	var announcements []Announcement

	for i, record := range records {
		fmt.Printf("  处理记录 %d/%d: %s\n", i+1, len(records), record.Title)

		// 清理标题（移除高亮标签）
		cleanTitle := cleanHTMLTags(record.Title)

		// 构建完整URL
		fullURL := buildFullURL(record.Linkurl)

		// 获取公告详情（主要为了获取采购人信息）
		publisher, err := extractPublisherFromDetail(client, fullURL)
		if err != nil {
			fmt.Printf("  警告: 提取采购人信息失败: %v\n", err)
			// 如果提取失败，尝试从API返回的内容中提取
			publisher = extractPublisherFromContent(record.Content)
		}

		// 验证是否是生态环境局的公告
		if isEcologyPublisher(publisher) {
			announcement := Announcement{
				Title:       cleanTitle,
				Publisher:   publisher,
				PublishDate: formatDate(record.Webdate),
				URL:         fullURL,
				Content:     cleanHTMLTags(record.Content),
				Source:      "API搜索",
			}
			announcements = append(announcements, announcement)
			fmt.Printf("  找到生态环境局公告: %s (发布单位: %s)\n", cleanTitle, publisher)
		} else {
			fmt.Printf("  跳过: 发布单位不匹配 - %s\n", publisher)
		}

		// 避免请求过快
		if i%5 == 0 && i > 0 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	return announcements
}

// 从详情页提取采购人信息
func extractPublisherFromDetail(client *http.Client, url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	setCommonHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 从HTML内容中提取采购人信息
	return extractPublisherFromHTML(string(body)), nil
}

// 从HTML内容中提取采购人信息
func extractPublisherFromHTML(htmlContent string) string {
	// 多种模式尝试提取采购人信息
	patterns := []string{
		// 采购人信息模式
		`采购人[信息]?[：:]\s*(?:名称[：:]?\s*)?([^<&\n\r]+?)(?:地址|联系方式|采购代理|项目联系|</)`,
		`采购人名称[：:]\s*([^<&\n\r]+?)(?:地址|联系方式|</)`,
		`招标人[：:]\s*(?:名称[：:]?\s*)?([^<&\n\r]+?)(?:地址|联系方式|</)`,
		`甲方[：:]\s*(?:名称[：:]?\s*)?([^<&\n\r]+?)(?:地址|联系方式|乙方|</)`,

		// 表格模式
		`<td[^>]*>\s*(?:采购人|招标人|甲方)\s*</td>\s*<td[^>]*>\s*([^<]+?)\s*</td>`,
		`<td[^>]*>\s*(?:采购人|招标人|甲方)名称\s*</td>\s*<td[^>]*>\s*([^<]+?)\s*</td>`,

		// 通用模式
		`名称[：:]\s*([^<&\n\r]+?生态环境局[^<&\n\r]*?)(?:地址|联系方式|</)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			publisher := strings.TrimSpace(matches[1])
			// 清理可能的HTML实体
			publisher = strings.ReplaceAll(publisher, "&nbsp;", " ")
			publisher = strings.ReplaceAll(publisher, "&#160;", " ")
			publisher = strings.ReplaceAll(publisher, "&amp;", "&")

			if publisher != "" && strings.Contains(publisher, "生态环境局") {
				return publisher
			}
		}
	}

	// 如果没有找到明确的采购人信息，尝试从内容中提取
	return extractPublisherFromContent(htmlContent)
}

// 从内容中提取采购人信息
func extractPublisherFromContent(content string) string {
	// 清理HTML标签
	cleanContent := cleanHTMLTags(content)

	// 查找包含"生态环境局"的句子
	sentences := strings.Split(cleanContent, "。")
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if strings.Contains(sentence, "生态环境局") {
			// 尝试提取单位名称
			if idx := strings.Index(sentence, "生态环境局"); idx > 0 {
				// 将字符串转换为rune切片以便正确处理中文字符
				runes := []rune(sentence)
				idxRune := strings.IndexRune(sentence, '生')
				if idxRune < 0 {
					continue
				}

				// 向前查找单位名称的开头
				start := idxRune - 1
				for start >= 0 && !isChinesePunctuation(runes[start]) {
					start--
				}
				if start < 0 {
					start = 0
				} else {
					start++ // 跳过标点符号
				}

				// 向后查找单位名称的结尾
				end := idxRune + len([]rune("生态环境局"))
				for end < len(runes) && !isChinesePunctuation(runes[end]) {
					end++
				}

				publisher := strings.TrimSpace(string(runes[start:end]))
				if publisher != "" {
					return publisher
				}
			}
		}
	}

	return "未知单位"
}

// 检查是否是生态环境局公告
func isEcologyPublisher(publisher string) bool {
	if publisher == "未知单位" || publisher == "" {
		return false
	}

	// 检查是否包含生态环境局
	return strings.Contains(publisher, "生态环境局")
}

// 清理HTML标签
func cleanHTMLTags(html string) string {
	// 移除高亮标签
	html = strings.ReplaceAll(html, "<em style='color:red'>", "")
	html = strings.ReplaceAll(html, "</em>", "")

	// 移除其他常见HTML标签
	re := regexp.MustCompile(`<[^>]*>`)
	html = re.ReplaceAllString(html, "")

	// 替换HTML实体
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	html = strings.ReplaceAll(html, "&#160;", " ")
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&quot;", "\"")

	return strings.TrimSpace(html)
}

// 格式化日期
func formatDate(dateStr string) string {
	// 尝试多种日期格式
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

// 检查是否是中文标点符号
func isChinesePunctuation(c rune) bool {
	punctuations := []rune{'，', '。', '；', '：', '？', '！', '、', '（', '）', '《', '》', '【', '】'}
	for _, p := range punctuations {
		if c == p {
			return true
		}
	}
	return false
}

// 设置API请求头
func setAPIHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Origin", "http://zfcg.szggzy.com:8081")
	req.Header.Set("Referer", "http://zfcg.szggzy.com:8081/")
}

// 设置通用请求头
func setCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "http://zfcg.szggzy.com:8081/")
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

// 保存结果到文件
func saveResultsToFile(announcements []Announcement) {
	if len(announcements) == 0 {
		return
	}

	// 保存为JSON文件
	jsonData, err := json.MarshalIndent(announcements, "", "  ")
	if err != nil {
		fmt.Printf("保存JSON文件失败: %v\n", err)
		return
	}

	filename := fmt.Sprintf("生态环境局公告_%s.json", time.Now().Format("20060102_150405"))
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
	} else {
		fmt.Printf("结果已保存到文件: %s\n", filename)
	}
}
