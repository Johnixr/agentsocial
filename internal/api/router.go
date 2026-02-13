package api

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"

	"agentsocial/internal/config"
	"agentsocial/internal/core"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter creates and configures the gin router with all routes and middleware.
func SetupRouter(db *sql.DB, cfg *config.Config, embClient *core.EmbeddingClient) *gin.Engine {
	router := gin.Default()

	// CORS middleware: allow all origins for development.
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// API v1 routes.
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication required).
		v1.POST("/agents/register", RegisterAgent(db, cfg, embClient))

		pub := v1.Group("/public")
		{
			pub.GET("/agents", ListPublicAgents(db))
			pub.GET("/agents/:id", GetPublicAgent(db))
			pub.GET("/tasks/:id", GetPublicTask(db))
			pub.GET("/stats", GetPublicStats(db))
		}

		// Authenticated routes.
		auth := v1.Group("")
		auth.Use(AuthMiddleware(db))
		{
			auth.GET("/agents/me", GetMe(db))
			auth.PUT("/agents/tasks/:taskId", UpdateTask(db, embClient))
			auth.POST("/scan", Scan(db, cfg, embClient))
			auth.POST("/conversations", CreateConversation(db))
			auth.GET("/conversations", ListConversations(db))
			auth.POST("/heartbeat", Heartbeat(db))
			auth.POST("/reports", CreateReport(db, cfg))
		}
	}

	// Serve static files for the SPA frontend (if built).
	webDist := filepath.Join("web", "dist")
	if info, err := os.Stat(webDist); err == nil && info.IsDir() {
		router.Static("/assets", filepath.Join(webDist, "assets"))

		// Serve index.html for the root and any unmatched routes (SPA fallback).
		router.NoRoute(func(c *gin.Context) {
			// If the request is for an API route, return 404.
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "not_found",
					"message": "API endpoint not found",
				})
				return
			}
			c.File(filepath.Join(webDist, "index.html"))
		})
	} else {
		router.NoRoute(func(c *gin.Context) {
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "not_found",
					"message": "API endpoint not found",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "AgentSocial API is running. Frontend not built yet.",
				"docs":    "See /api/v1/public/stats for platform statistics.",
			})
		})
	}

	return router
}
