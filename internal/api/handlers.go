package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ieasydevops/demo-scrapy/internal/database"
	"github.com/ieasydevops/demo-scrapy/internal/models"
	"github.com/ieasydevops/demo-scrapy/internal/scheduler"
)

// GetWebPages 获取网页列表
// @Summary      获取网页列表
// @Description  获取所有监控网页的列表
// @Tags         网页管理
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.WebPage
// @Failure      500  {object}  map[string]string
// @Router       /web-pages [get]
func GetWebPages(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, url, name FROM web_pages")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var pages []models.WebPage
	for rows.Next() {
		var page models.WebPage
		if err := rows.Scan(&page.ID, &page.URL, &page.Name); err != nil {
			continue
		}
		pages = append(pages, page)
	}

	c.JSON(http.StatusOK, pages)
}

// CreateWebPage 创建网页
// @Summary      创建网页
// @Description  添加一个新的监控网页
// @Tags         网页管理
// @Accept       json
// @Produce      json
// @Param        page  body      models.WebPage  true  "网页信息"
// @Success      200   {object}  models.WebPage
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /web-pages [post]
func CreateWebPage(c *gin.Context) {
	var page models.WebPage
	if err := c.ShouldBindJSON(&page); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := database.DB.Exec("INSERT INTO web_pages (url, name) VALUES (?, ?)", page.URL, page.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	page.ID = int(id)

	c.JSON(http.StatusOK, page)
}

// UpdateWebPage 更新网页
// @Summary      更新网页
// @Description  更新指定ID的网页信息
// @Tags         网页管理
// @Accept       json
// @Produce      json
// @Param        id    path      int            true  "网页ID"
// @Param        page  body      models.WebPage true  "网页信息"
// @Success      200   {object}  models.WebPage
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /web-pages/{id} [put]
func UpdateWebPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var page models.WebPage
	if err := c.ShouldBindJSON(&page); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec("UPDATE web_pages SET url = ?, name = ? WHERE id = ?", page.URL, page.Name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	page.ID = id
	c.JSON(http.StatusOK, page)
}

// DeleteWebPage 删除网页
// @Summary      删除网页
// @Description  删除指定ID的网页
// @Tags         网页管理
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "网页ID"
// @Success      200 {object}  map[string]string
// @Failure      500 {object}   map[string]string
// @Router       /web-pages/{id} [delete]
func DeleteWebPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := database.DB.Exec("DELETE FROM web_pages WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetKeywords 获取关键字列表
// @Summary      获取关键字列表
// @Description  获取所有监控关键字
// @Tags         关键字管理
// @Accept       json
// @Produce      json
// @Success      200 {array}   models.Keyword
// @Failure      500 {object} map[string]string
// @Router       /keywords [get]
func GetKeywords(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, keyword FROM keywords")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var keywords []models.Keyword
	for rows.Next() {
		var keyword models.Keyword
		if err := rows.Scan(&keyword.ID, &keyword.Keyword); err != nil {
			continue
		}
		keywords = append(keywords, keyword)
	}

	c.JSON(http.StatusOK, keywords)
}

// CreateKeyword 创建关键字
// @Summary      创建关键字
// @Description  添加一个新的监控关键字，并立即重新加载任务
// @Tags         关键字管理
// @Accept       json
// @Produce      json
// @Param        keyword  body      models.Keyword  true  "关键字信息"
// @Success      200      {object}  models.Keyword
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /keywords [post]
func CreateKeyword(c *gin.Context) {
	var keyword models.Keyword
	if err := c.ShouldBindJSON(&keyword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := database.DB.Exec("INSERT INTO keywords (keyword) VALUES (?)", keyword.Keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	keyword.ID = int(id)

	go scheduler.ReloadTasks()

	c.JSON(http.StatusOK, keyword)
}

// DeleteKeyword 删除关键字
// @Summary      删除关键字
// @Description  删除指定ID的关键字，并立即重新加载任务
// @Tags         关键字管理
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "关键字ID"
// @Success      200 {object}  map[string]string
// @Failure      500 {object}  map[string]string
// @Router       /keywords/{id} [delete]
func DeleteKeyword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := database.DB.Exec("DELETE FROM keywords WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go scheduler.ReloadTasks()

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetPushConfig 获取推送配置
// @Summary      获取推送配置
// @Description  获取邮件推送配置信息
// @Tags         推送配置
// @Accept       json
// @Produce      json
// @Success      200 {object} models.PushConfig
// @Router       /push-config [get]
func GetPushConfig(c *gin.Context) {
	var config models.PushConfig
	err := database.DB.QueryRow("SELECT id, email, push_time FROM push_config LIMIT 1").Scan(&config.ID, &config.Email, &config.PushTime)
	if err != nil {
		c.JSON(http.StatusOK, models.PushConfig{Email: "403608355@qq.com", PushTime: "17"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdatePushConfig 更新推送配置
// @Summary      更新推送配置
// @Description  更新邮件推送配置（邮箱和推送时间）
// @Tags         推送配置
// @Accept       json
// @Produce      json
// @Param        config  body      models.PushConfig  true  "推送配置"
// @Success      200     {object}  models.PushConfig
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /push-config [put]
func UpdatePushConfig(c *gin.Context) {
	var config models.PushConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM push_config").Scan(&exists)

	if exists == 0 {
		_, err := database.DB.Exec("INSERT INTO push_config (email, push_time) VALUES (?, ?)", config.Email, config.PushTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		_, err := database.DB.Exec("UPDATE push_config SET email = ?, push_time = ?", config.Email, config.PushTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	scheduler.ReloadTasks()

	c.JSON(http.StatusOK, config)
}
