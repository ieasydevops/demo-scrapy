package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ieasydevops/demo-scrapy/internal/database"
	"github.com/ieasydevops/demo-scrapy/internal/models"
	"github.com/ieasydevops/demo-scrapy/internal/scheduler"
)

// GetSubscribeConfig 获取订阅配置列表
// @Summary      获取订阅配置列表
// @Description  获取所有订阅用户邮箱列表
// @Tags         订阅配置管理
// @Accept       json
// @Produce      json
// @Success      200 {array} models.SubscribeConfig
// @Failure      500 {object} map[string]string
// @Router       /subscribe-config [get]
func GetSubscribeConfig(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, email, push_time, created_at FROM subscribe_config ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var configs []models.SubscribeConfig
	for rows.Next() {
		var config models.SubscribeConfig
		if err := rows.Scan(&config.ID, &config.Email, &config.PushTime, &config.CreatedAt); err != nil {
			continue
		}
		configs = append(configs, config)
	}

	c.JSON(http.StatusOK, configs)
}

// CreateSubscribeConfig 创建订阅配置
// @Summary      创建订阅配置
// @Description  添加新的订阅用户邮箱
// @Tags         订阅配置管理
// @Accept       json
// @Produce      json
// @Param        config  body      models.SubscribeConfig  true  "订阅配置"
// @Success      200     {object}  models.SubscribeConfig
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /subscribe-config [post]
func CreateSubscribeConfig(c *gin.Context) {
	var config models.SubscribeConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO subscribe_config (email, push_time) VALUES (?, ?)",
		config.Email, config.PushTime,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	config.ID = int(id)
	
	go scheduler.ReloadTasks()
	
	c.JSON(http.StatusOK, config)
}

// UpdateSubscribeConfig 更新订阅配置
// @Summary      更新订阅配置
// @Description  更新指定ID的订阅配置
// @Tags         订阅配置管理
// @Accept       json
// @Produce      json
// @Param        id      path      int                    true  "配置ID"
// @Param        config  body      models.SubscribeConfig true  "订阅配置"
// @Success      200     {object}  models.SubscribeConfig
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /subscribe-config/{id} [put]
func UpdateSubscribeConfig(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var config models.SubscribeConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec(
		"UPDATE subscribe_config SET email = ?, push_time = ? WHERE id = ?",
		config.Email, config.PushTime, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config.ID = id
	
	go scheduler.ReloadTasks()
	
	c.JSON(http.StatusOK, config)
}

// DeleteSubscribeConfig 删除订阅配置
// @Summary      删除订阅配置
// @Description  删除指定ID的订阅配置
// @Tags         订阅配置管理
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "配置ID"
// @Success      200 {object}  map[string]string
// @Failure      500 {object}  map[string]string
// @Router       /subscribe-config/{id} [delete]
func DeleteSubscribeConfig(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := database.DB.Exec("DELETE FROM subscribe_config WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go scheduler.ReloadTasks()

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
