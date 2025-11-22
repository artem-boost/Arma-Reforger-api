package main

import (
	"arma-reforger-api/handlers"
	"arma-reforger-api/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"

	"github.com/gin-gonic/gin"
)

var (
	config *models.Config
)

func loadConfig() error {
	file, err := os.Open("config.json")
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("failed to decode config: %v", err)
	}

	models.SetConfig(config)
	return nil
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, x-client-id, x-client-secret")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	router.GET("/game-identity/api/v1.0/health", handlers.APIStatusOk)
	// Auth routes
	auth := router.Group("/game-identity/api/v1.1/identities/reforger")
	{
		auth.POST("/auth", handlers.AuthHandler)
	}

	// Game API routes
	gameAPI := router.Group("/game-api")
	{
		gameAPI.GET("/health", handlers.APIStatusOk)
		// S2S API
		s2s := gameAPI.Group("/s2s-api/v1.0/lobby")
		{
			s2s.POST("/rooms/acceptPlayer", handlers.AcceptPlayerHandler)
			s2s.POST("/dedicatedServers/registerUnmanagedServer", handlers.RegisterUnmanagedServerHandler)
			s2s.POST("/rooms/register", handlers.RegisterRoomHandler)
			s2s.POST("/rooms/remove", handlers.RemoveServerHandler)
			s2s.POST("/dedicatedServers/heartBeat", handlers.HeartBeatHandler)
			s2s.POST("/rooms/listActiveBans", handlers.ListActiveBansHandler)
			s2s.POST("/rooms/removePlayer", handlers.RemovePlayerHandler)
			s2s.POST("/rooms/createBan", handlers.CreateBanHandler)
			s2s.POST("/rooms/removeBans", handlers.RemoveBansHandler)
			s2s.POST("/sendTdEvents", handlers.SendTdEventsHandler)
		}

		// Public API
		api := gameAPI.Group("/api/v1.0")
		{
			room := api.Group("/lobby/rooms")
			{
				room.POST("/search", handlers.SearchServersHandler)
				room.POST("/getRoomsByIds", handlers.GetRoomsByIDsHandler)
				room.POST("/join", handlers.JoinRoomHandler)
				room.POST("/verifyPassword", handlers.VerifyPasswordHandler)
				room.POST("/register", handlers.RegisterRoomHandler)
				room.POST("/heartBeat", handlers.RoomHeartBeatHandler)
				room.POST("/remove", handlers.RemoveRoomHandler)
				room.POST("/acceptPlayer", handlers.AcceptPlayerHandler)
				room.POST("/listPlayers", handlers.ListPlayersHandler)
			}
			api.POST("/lobby/dedicatedServers/createOwnerToken", handlers.CreateOwnerTokenHandler)
			api.POST("/lobby/getPingSites", handlers.GetPingSitesHandler)
			api.POST("/session/login", handlers.SessionLoginHandler)
			api.POST("/blockList/listBlocked", handlers.BlockListHandler)
			api.POST("/sendTdEvents", handlers.SendTdEventsHandler)
			api.GET("/world", handlers.WorldHandler)
			api.GET("/dummy", handlers.DummyHandler)
		}
	}

	// Utility routes
	router.POST("/ping", handlers.PingHandler)
	router.GET("/api/news", handlers.NewsHandler)
	router.POST("/GetWorkshopToken", handlers.GetWorkshopTokenHandler)
	router.POST("/getLobby", handlers.GetLobbyHandler)

	// User API routes
	userAPI := router.Group("/user/api/v2.0")
	{
		userAPI.POST("/game-identity/token", handlers.GetUserTokenProxyHandler)
	}

	return router
}

func main() {
	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	if err := models.InitDatabase(config.DB.Path); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize ticket manager

	// Start cleanup goroutine for old servers
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := models.CleanupOldServers(); err != nil {
				log.Printf("Failed to cleanup old servers: %v", err)
			}
		}
	}()

	// Setup and start server
	router := setupRouter()

	log.Printf("Server starting on port %d", config.API.PORT)
	if err := router.Run(fmt.Sprintf(":%d", config.API.PORT)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
