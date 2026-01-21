package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ieasydevops/demo-scrapy/internal/database"
	"github.com/ieasydevops/demo-scrapy/internal/models"
)

// GetAnnouncements 获取公告列表
// @Summary      获取公告列表
// @Description  获取采购信息动态，支持时间排序和搜索
// @Tags         采购信息动态
// @Accept       json
// @Produce      json
// @Param        keyword  query     string  false  "搜索关键字"
// @Param        order    query     string  false  "排序方式: desc(降序) 或 asc(升序)" default(desc)
// @Param        page     query     int     false  "页码" default(1)
// @Param        pageSize query     int     false  "每页数量" default(20)
// @Success      200      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]string
// @Router       /announcements [get]
func GetAnnouncements(c *gin.Context) {
	keyword := c.Query("keyword")
	order := c.DefaultQuery("order", "desc")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	var query strings.Builder
	query.WriteString(`
		SELECT a.id, a.title, a.url, a.publish_date, a.content, a.created_at,
		       a.web_page_id, wp.name as web_page_name, a.publisher
		FROM announcements a
		LEFT JOIN web_pages wp ON a.web_page_id = wp.id
		WHERE 1=1
	`)

	args := []interface{}{}
	if keyword != "" {
		query.WriteString(" AND (a.title LIKE ? OR a.content LIKE ?)")
		keywordPattern := "%" + keyword + "%"
		args = append(args, keywordPattern, keywordPattern)
	}

	query.WriteString(" ORDER BY a.created_at " + strings.ToUpper(order))

	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM announcements a
		LEFT JOIN web_pages wp ON a.web_page_id = wp.id
		WHERE 1=1
	`
	countArgs := []interface{}{}
	if keyword != "" {
		countQuery += " AND (a.title LIKE ? OR a.content LIKE ?)"
		keywordPattern := "%" + keyword + "%"
		countArgs = append(countArgs, keywordPattern, keywordPattern)
	}
	err := database.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pageInt := 1
	pageSizeInt := 20
	if p, err := strconv.Atoi(page); err == nil && p > 0 {
		pageInt = p
	}
	if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
		pageSizeInt = ps
	}

	offset := (pageInt - 1) * pageSizeInt
	query.WriteString(" LIMIT ? OFFSET ?")
	args = append(args, pageSizeInt, offset)

	rows, err := database.DB.Query(query.String(), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var announcements []models.Announcement
	for rows.Next() {
		var ann models.Announcement
		var webPageID sql.NullInt64
		var webPageName sql.NullString
		var publisher sql.NullString
		err := rows.Scan(&ann.ID, &ann.Title, &ann.URL, &ann.PublishDate, &ann.Content,
			&ann.CreatedAt, &webPageID, &webPageName, &publisher)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("扫描数据失败: %v", err)})
			return
		}
		if webPageID.Valid {
			ann.WebPageID = int(webPageID.Int64)
		}
		if webPageName.Valid {
			ann.WebPageName = webPageName.String
		}
		if publisher.Valid {
			ann.Publisher = publisher.String
		}
		announcements = append(announcements, ann)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("遍历数据失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       announcements,
		"total":      total,
		"page":       pageInt,
		"page_size":  pageSizeInt,
		"total_page": (total + pageSizeInt - 1) / pageSizeInt,
	})
}
