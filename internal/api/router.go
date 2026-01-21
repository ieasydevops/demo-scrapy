package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		api.GET("/web-pages", GetWebPages)
		api.POST("/web-pages", CreateWebPage)
		api.PUT("/web-pages/:id", UpdateWebPage)
		api.DELETE("/web-pages/:id", DeleteWebPage)

		api.GET("/keywords", GetKeywords)
		api.POST("/keywords", CreateKeyword)
		api.DELETE("/keywords/:id", DeleteKeyword)

		api.GET("/monitor-config", GetMonitorConfig)
		api.POST("/monitor-config", CreateMonitorConfig)
		api.PUT("/monitor-config/:id", UpdateMonitorConfig)
		api.DELETE("/monitor-config/:id", DeleteMonitorConfig)

		api.GET("/subscribe-config", GetSubscribeConfig)
		api.POST("/subscribe-config", CreateSubscribeConfig)
		api.PUT("/subscribe-config/:id", UpdateSubscribeConfig)
		api.DELETE("/subscribe-config/:id", DeleteSubscribeConfig)

		api.GET("/announcements", GetAnnouncements)

		api.GET("/push-config", GetPushConfig)
		api.PUT("/push-config", UpdatePushConfig)
	}

	return r
}
