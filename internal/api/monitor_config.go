package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ieasydevops/demo-scrapy/internal/database"
	"github.com/ieasydevops/demo-scrapy/internal/models"
	"github.com/ieasydevops/demo-scrapy/internal/scheduler"
)

// GetMonitorConfig 获取监控配置
// @Summary      获取监控配置
// @Description  获取所有监控配置列表
// @Tags         监控配置管理
// @Accept       json
// @Produce      json
// @Success      200 {array} models.MonitorConfig
// @Failure      500 {object} map[string]string
// @Router       /monitor-config [get]
func GetMonitorConfig(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT mc.id, mc.web_page_id, mc.crawl_time, mc.crawl_freq, mc.keywords, 
		       mc.created_at, mc.updated_at, wp.name as web_page_name
		FROM monitor_config mc
		LEFT JOIN web_pages wp ON mc.web_page_id = wp.id
		ORDER BY mc.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var configs []map[string]interface{}
	for rows.Next() {
		var config models.MonitorConfig
		var webPageName string
		err := rows.Scan(&config.ID, &config.WebPageID, &config.CrawlTime, &config.CrawlFreq,
			&config.Keywords, &config.CreatedAt, &config.UpdatedAt, &webPageName)
		if err != nil {
			continue
		}
		configs = append(configs, map[string]interface{}{
			"id":            config.ID,
			"web_page_id":   config.WebPageID,
			"web_page_name": webPageName,
			"crawl_time":    config.CrawlTime,
			"crawl_freq":    config.CrawlFreq,
			"keywords":      strings.Split(config.Keywords, ","),
			"created_at":    config.CreatedAt,
			"updated_at":    config.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, configs)
}

// CreateMonitorConfig 创建监控配置
// @Summary      创建监控配置
// @Description  创建新的监控配置
// @Tags         监控配置管理
// @Accept       json
// @Produce      json
// @Param        config  body      object  true  "监控配置"
// @Success      200     {object}  map[string]interface{}
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /monitor-config [post]
func CreateMonitorConfig(c *gin.Context) {
	var req struct {
		WebPageID int      `json:"web_page_id" binding:"required"`
		CrawlTime string   `json:"crawl_time" binding:"required"`
		CrawlFreq string   `json:"crawl_freq" binding:"required"`
		Keywords  []string `json:"keywords" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keywordsStr := strings.Join(req.Keywords, ",")
	result, err := database.DB.Exec(
		"INSERT INTO monitor_config (web_page_id, crawl_time, crawl_freq, keywords) VALUES (?, ?, ?, ?)",
		req.WebPageID, req.CrawlTime, req.CrawlFreq, keywordsStr,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	
	go scheduler.ReloadTasks()
	
	c.JSON(http.StatusOK, gin.H{"id": id, "message": "created"})
}

// UpdateMonitorConfig 更新监控配置
// @Summary      更新监控配置
// @Description  更新指定ID的监控配置
// @Tags         监控配置管理
// @Accept       json
// @Produce      json
// @Param        id      path      int     true  "配置ID"
// @Param        config  body      object  true  "监控配置"
// @Success      200     {object}  map[string]string
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /monitor-config/{id} [put]
func UpdateMonitorConfig(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		WebPageID int      `json:"web_page_id"`
		CrawlTime string   `json:"crawl_time"`
		CrawlFreq string   `json:"crawl_freq"`
		Keywords  []string `json:"keywords"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keywordsStr := strings.Join(req.Keywords, ",")
	_, err := database.DB.Exec(
		"UPDATE monitor_config SET web_page_id = ?, crawl_time = ?, crawl_freq = ?, keywords = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		req.WebPageID, req.CrawlTime, req.CrawlFreq, keywordsStr, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	scheduler.ReloadTasks()
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteMonitorConfig 删除监控配置
// @Summary      删除监控配置
// @Description  删除指定ID的监控配置
// @Tags         监控配置管理
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "配置ID"
// @Success      200 {object}  map[string]string
// @Failure      500 {object}  map[string]string
// @Router       /monitor-config/{id} [delete]
func DeleteMonitorConfig(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := database.DB.Exec("DELETE FROM monitor_config WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go scheduler.ReloadTasks()

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
