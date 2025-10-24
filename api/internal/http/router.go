package http

import (
	"github.com/gin-contrib/cors"

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

	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,  // ONLY FOR TESTING!
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Initialize shared session store for token revocation
	sessionStore := session.NewInMemorySessionStore()

	// Initialize handlers (passing shared session store to auth handler)
	authHandler := handlers.NewAuthHandler(db, cfg, sessionStore)
	videoHandler := handlers.NewVideoHandler(db, cfg)
	voteHandler := handlers.NewVoteHandler(db)
	rankingHandler := handlers.NewRankingHandler(db)
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
			auth.GET("/profile", authMiddleware, authHandler.Profile)
		}

		videos := api.Group("/videos")
		{
			videos.POST("/upload", authMiddleware, videoHandler.UploadVideo)
			videos.GET("/", authMiddleware, videoHandler.GetUserVideos)
			videos.GET("/:video_id", authMiddleware, videoHandler.GetVideo)
			videos.DELETE("/:video_id", authMiddleware, videoHandler.DeleteVideo)
		}

		public := api.Group("/public")
		{
			public.GET("/videos", videoHandler.GetPublicVideos)

			// Vote endpoints require authentication
			public.POST("/videos/:video_id/vote", authMiddleware, voteHandler.VoteForVideo)
			public.DELETE("/videos/:video_id/vote", authMiddleware, voteHandler.UnvoteForVideo)
			public.GET("/videos/:video_id/stream", videoHandler.StreamVideo)

			// Rankings endpoints (no authentication required)
			public.GET("/rankings", rankingHandler.GetPlayerRankings)
		}

		router.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Backend API running"})
		})

	}

	return router
}
