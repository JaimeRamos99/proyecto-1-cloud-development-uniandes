package http

import (
	"proyecto1/root/internal/auth"
	"proyecto1/root/internal/config"
	"proyecto1/root/internal/database"
	"proyecto1/root/internal/http/handlers"
	"proyecto1/root/internal/http/middlewares"
	"proyecto1/root/internal/http/session"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, db *database.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// CORS is handled entirely by nginx reverse proxy
	// All requests come through nginx, so no CORS configuration needed here

	// Initialize shared session store for token revocation
	sessionStore := session.NewInMemorySessionStore()
	
	// Initialize handlers (passing shared session store to auth handler)
	authHandler := handlers.NewAuthHandler(db, cfg, sessionStore)
	videoHandler := handlers.NewVideoHandler(db, cfg)
	healthHandler := handlers.NewHealthHandler(db, videoHandler)

	// Initialize auth middleware with shared session store
	tokenManager := &auth.TokenManager{
		Secret: []byte(cfg.JWT.Secret),
		Issuer: cfg.JWT.Issuer,
	}
	authMiddleware := middlewares.AuthMiddleware(*tokenManager, sessionStore.IsTokenRevoked)

	api := router.Group("/api")
	{
		// Health check endpoint
		api.GET("/health", healthHandler.Health)

		auth := api.Group("/auth")
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		videos := api.Group("/videos")
		{
			videos.POST("/upload", authMiddleware, videoHandler.UploadVideo)
			videos.GET("/", authMiddleware, videoHandler.GetUserVideos)
			videos.GET("/:video_id", authMiddleware, videoHandler.GetVideo)
			videos.DELETE("/:video_id", authMiddleware, videoHandler.DeleteVideo)
		}
	}
	
	return router
}