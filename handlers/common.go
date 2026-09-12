package handlers

import (
	"arma-reforger-api/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "pong"})
}

func NewsHandler(c *gin.Context) {
	news := gin.H{
		"items": []gin.H{
			{
				"date":     "08 Августа 2024",
				"excerpt":  "Начало тестирования нового API",
				"category": "Development",
				"slug":     "update-august-08-2024",
				"title":    "1.2.0.102 Update",
				"coverImage": gin.H{
					"src": "https://cms-cdn.bistudio.com/cms-static--reforger/images/08f81027-c102-4a4f-b9e7-fe12e8e6e8c2-NEWS%201280x720.jpg",
				},
				"fullUrl": "https://youtu.be/dQw4w9WgXcQ?si=zY34xDJ__8psHZ3i",
			},
		},
	}
	c.JSON(http.StatusOK, news)
}

func BlockListHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
		"blockList": gin.H{
			"entries":    []interface{}{},
			"totalCount": 0,
			"page": gin.H{
				"offset": 0,
				"limit":  16,
			},
		},
	})
}

func ListPlayersHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"connectedPlayers": []interface{}{},
		"queuePlayers":     []interface{}{},
	})
}

func SendTdEventsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func GetPingSitesHandler(c *gin.Context) {
	pingSites := gin.H{
		"pingSites": []gin.H{
			{
				"id":        "frankfurt",
				"address":   "ping-location-de.nitrado.net",
				"ipAddress": "31.214.130.69",
				"location": gin.H{
					"latitude":  51.29930114746094,
					"longitude": 9.491000175476074,
				},
				"mappedRegions": []string{"eu-ffm"},
			},
			// Add other ping sites as needed
		},
	}
	c.JSON(http.StatusOK, pingSites)
}

func GetLobbyHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"rooms": []interface{}{}})
}

func WorldHandler(c *gin.Context) {
	world := gin.H{
		"version": "BanSettings",
		"gameData": gin.H{
			"BanSettings": gin.H{
				"m_sDesc": "Ban Logic Cleansweep",
				"m_BanSettings": gin.H{
					"m_fScoreThreshold":          10.0,
					"m_fScoreDecreasePerMinute":  0.2,
					"m_fScoreMultiplier":         0.2,
					"m_fAccelerationMin":         1.0,
					"m_fAccelerationMax":         6.0,
					"m_fBanEvaluationLight":      0.8,
					"m_fBanEvaluationHeavy":      1.0,
					"m_fCrimePtFriendKill":       1.0,
					"m_fCrimePtTeamKill":         0.7,
					"m_fQualityTimeTemp":         1.0,
					"m_bVotingSuggestionEnabled": 0,
				},
			},
		},
	}
	c.JSON(http.StatusOK, world)
}

func DummyHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func SessionLoginHandler(c *gin.Context) {
	data := gin.H{
		"userProfile": gin.H{
			"userId":      "1bde4705-34fd-489d-a7fe-93f3a2f5aefc",
			"username":    "",
			"renameCount": -1,
			"currencies": gin.H{
				"HardCurrency": 0,
				"SoftCurrency": 0,
			},
			"countryCode":        "UA",
			"overallPlayTime":    66230,
			"tester":             false,
			"isDeveloperAccount": false,
			"rentedServers": gin.H{
				"entries":      []interface{}{},
				"visitedGames": []interface{}{},
			},
		},
		"worldVersion":             "BanSettings",
		"ipAddress":                "91.219.235.155",
		"pendingMicroTransactions": []interface{}{},
		"compatibleGameVersions":   []string{"1.1.0.42"},
		"notifications":            []interface{}{},
		"sessionId":                "4105b12d-6873-4f8e-9dd3-36d638fdc455",
	}
	c.JSON(http.StatusOK, data)
}

// ==CONFIG HANDLER==

// Описание структуры JSON ответа
type ConfigOverview struct {
	Tag            *string `json:"tag"`
	Status         string  `json:"status"`
	Message        string  `json:"message"`
	HealthCheckUrl *string `json:"healthCheckUrl"`
}

type ConfigProperty struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Value *string `json:"value,omitempty"` // omitempty уберет поле, если оно nil (как в URI_StarRiver)
}

type ConfigService struct {
	Tag            string  `json:"tag"`
	Status         string  `json:"status"`
	Message        string  `json:"message"`
	HealthCheckUrl *string `json:"healthCheckUrl"`
}

type GameConfigResponse struct {
	Name       string           `json:"name"`
	Version    string           `json:"version"`
	Env        string           `json:"env"`
	Overview   ConfigOverview   `json:"overview"`
	Properties []ConfigProperty `json:"properties"`
	Services   []ConfigService  `json:"services"`
	CreatedAt  string           `json:"createdAt"`
}

// Вспомогательная функция для указателей на строки
func strPtr(s string) *string {
	return &s
}

func GameConfigHandler(c *gin.Context) {
	baseURL := models.GetConfig().API.PublishHost
	// 1. Получаем путь после /game-config/
	// Например: "/api/v1/reforger/1.8.0/13/list"
	fullPath := c.Param("path")

	// 2. Извлекаем версию из пути
	gameVersion := "1.8.0" // Устанавливаем дефолтное значение на случай непредвиденного URL
	parts := strings.Split(fullPath, "/")

	// Ищем "reforger" и берем следующий за ним сегмент
	for i, part := range parts {
		if part == "reforger" && i+1 < len(parts) {
			gameVersion = parts[i+1]
			break
		}
	}
	// Формируем правильный ответ, соответствующий дампу
	response := GameConfigResponse{
		Name:    "Reforger",
		Version: gameVersion,
		Env:     "Production",
		Overview: ConfigOverview{
			Tag:            nil,
			Status:         "ok",
			Message:        "",
			HealthCheckUrl: nil,
		},
		Properties: []ConfigProperty{
			{Name: "Link_Bohemia", Type: "string", Value: strPtr("https://www.bohemia.net")},
			// ... (здесь можно перечислить все остальные ссылки, если они нужны клиенту) ...

			// Самые важные API эндпоинты
			{Name: "URI_ClientLobbyApi", Type: "string", Value: strPtr(baseURL + "/game-api/api/v1.0/lobby/")},
			{Name: "URI_GameApi", Type: "string", Value: strPtr(baseURL + "/game-api/api/v1.0/")},
			{Name: "URI_GameApiS2S", Type: "string", Value: strPtr(baseURL + "/game-api/s2s-api/v1.0/")},
			{Name: "URI_IdentityApi", Type: "string", Value: strPtr(baseURL + "/game-identity/")},
			{Name: "URI_LobbyApi", Type: "string", Value: strPtr(baseURL + "/game-api/s2s-api/v1.0/lobby/")},
			{Name: "URI_StorageApi", Type: "string", Value: strPtr(baseURL + "/storage/api/v4.0/")},
			{Name: "URI_UserApi", Type: "string", Value: strPtr(baseURL + "/user/api/v2.0/")},
			{Name: "URI_GroupApi", Type: "string", Value: strPtr(baseURL + "/group/api/v2.0/")},

			// Античит и аналитика
			{Name: "URI_STSContext", Type: "string", Value: strPtr(baseURL + "/game-identity/api/v1.1/nitrado/steel-shield/reforger/")},
			{Name: "Opt_Analytics", Type: "string", Value: strPtr("TreasureData")},

			// Пример параметра без value (останется только name и type)
			{Name: "URI_StarRiver", Type: "string"},
		},
		Services: []ConfigService{
			{
				Tag:            "game-api",
				Status:         "ok",
				Message:        "",
				HealthCheckUrl: strPtr(baseURL + "/game-api/health"),
			},
			{
				Tag:            "game-identity",
				Status:         "ok",
				Message:        "",
				HealthCheckUrl: strPtr(baseURL + "/game-identity/api/v1.0/health"),
			},
			{
				Tag:            "storage-api-v2",
				Status:         "ok",
				Message:        "",
				HealthCheckUrl: nil,
			},
		},
		// Ставим текущее время в нужном формате
		CreatedAt: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	}

	c.JSON(http.StatusOK, response)
}
