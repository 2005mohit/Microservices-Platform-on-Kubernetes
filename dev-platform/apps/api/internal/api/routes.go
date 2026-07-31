package api

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(domain string) *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	h := &Handler{domain: domain}
	ws := &WSHandler{}

	api := r.Group("/api/v1")
	{
		api.GET("/projects", h.ListProjects)
		api.POST("/projects", h.CreateProject)
		api.GET("/projects/:id", h.GetProject)
		api.DELETE("/projects/:id", h.DeleteProject)

		api.POST("/projects/:id/deploy", h.TriggerDeploy)
		api.GET("/deployments/:id", h.GetDeployment)
		api.GET("/projects/:id/deployments", h.ListDeployments)
	}

	r.GET("/ws/deployments/:id", ws.HandleWebSocket)

		api.POST("/webhooks/github", h.HandleGithubWebhook)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
