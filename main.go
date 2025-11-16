package main

import (
	"arma-reforger-api/handlers"
	"arma-reforger-api/models"
	"arma-reforger-api/utils"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"

	"github.com/gin-gonic/gin"
)

var (
	config        *models.Config
	ticketManager *utils.TicketManager
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
	utils.SetServersAPI(config.ServersAPI)
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
			rooms := api.Group("/lobby/rooms")
			{
				rooms.POST("/search", handlers.SearchServersHandler)
				rooms.POST("/getRoomsByIds", handlers.GetRoomsByIDsHandler)
				rooms.POST("/join", handlers.JoinRoomHandler)
				rooms.POST("/verifyPassword", handlers.VerifyPasswordHandler)
				rooms.POST("/register", handlers.RegisterRoomHandler)
				rooms.POST("/heartBeat", handlers.RoomHeartBeatHandler)
				rooms.POST("/remove", handlers.RemoveRoomHandler)
				rooms.POST("/acceptPlayer", handlers.AcceptPlayerHandler)
				rooms.POST("/listPlayers", handlers.ListPlayersHandler)
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

	// Workshop API
	workshop := router.Group("/workshop-api/api/v3.0")
	{
		workshop.POST("/assets/list", handlers.WorkshopListHandler)
	}

	// Utility routes
	router.POST("/ping", handlers.PingHandler)
	router.GET("/api/news", handlers.NewsHandler)
	router.POST("/GetWorkshopToken", handlers.GetWorkshopTokenHandler)
	router.POST("/getLobby", handlers.GetLobbyHandler)

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
	ticketManager = utils.NewTicketManager(config)
	go ticketManager.FetchTicketPeriodically()

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
	go func() {
		utils.CreateRooms(utils.GetRooms())
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			utils.CreateRooms(utils.GetRooms())
		}
	}()

	// Setup and start server
	router := setupRouter()

	log.Printf("Server starting on port %d", config.API.PORT)
	if err := router.Run(fmt.Sprintf(":%d", config.API.PORT)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
